package e2e_test

import (
	"fmt"
	"os"

	exec_util "github.com/appscode/kutil/tools/exec"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/elasticsearch/test/e2e/framework"
	"github.com/kubedb/elasticsearch/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	store "kmodules.xyz/objectstore-api/api/v1"
)

const (
	S3_BUCKET_NAME       = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME      = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME = "SWIFT_CONTAINER_NAME"
)

var _ = Describe("Elasticsearch", func() {
	var (
		err                      error
		f                        *framework.Invocation
		elasticsearch            *api.Elasticsearch
		garbageElasticsearch     *api.ElasticsearchList
		elasticsearchVersion     *catalog.ElasticsearchVersion
		snapshot                 *api.Snapshot
		snapshotPVC              *core.PersistentVolumeClaim
		secret                   *core.Secret
		skipMessage              string
		skipSnapshotDataChecking bool
	)

	BeforeEach(func() {
		f = root.Invoke()
		elasticsearch = f.CombinedElasticsearch()
		elasticsearchVersion = f.ElasticsearchVersion()
		garbageElasticsearch = new(api.ElasticsearchList)
		snapshot = f.Snapshot()
		secret = new(core.Secret)
		skipMessage = ""
		skipSnapshotDataChecking = true
	})

	var createAndWaitForRunning = func() {
		By("Create ElasticsearchVersion: " + elasticsearchVersion.Name)
		err = f.CreateElasticsearchVersion(elasticsearchVersion)
		Expect(err).NotTo(HaveOccurred())

		By("Create Elasticsearch: " + elasticsearch.Name)
		err = f.CreateElasticsearch(elasticsearch)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running elasticsearch")
		f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())
	}

	var deleteTestResource = func() {
		if elasticsearch == nil {
			Skip("Skipping")
		}

		By("Check if elasticsearch " + elasticsearch.Name + " exists.")
		es, err := f.GetElasticsearch(elasticsearch.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Elasticsearch was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete elasticsearch: " + elasticsearch.Name)
		err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Elasticsearch was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if es.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for elasticsearch to be paused")
			f.EventuallyDormantDatabaseStatus(elasticsearch.ObjectMeta).Should(matcher.HavePaused())

			By("Set DormantDatabase Spec.WipeOut to true")
			_, err = f.PatchDormantDatabase(elasticsearch.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(elasticsearch.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for elasticsearch resources to be wipedOut")
		f.EventuallyWipedOut(elasticsearch.ObjectMeta).Should(Succeed())
	}

	AfterEach(func() {
		// Delete test resource
		deleteTestResource()

		for _, es := range garbageElasticsearch.Items {
			*elasticsearch = es
			// Delete test resource
			deleteTestResource()
		}

		if !skipSnapshotDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
		}

		if secret != nil {
			f.DeleteSecret(secret.ObjectMeta)
		}

		if snapshotPVC != nil {
			err := f.DeletePersistentVolumeClaim(snapshotPVC.ObjectMeta)
			if err != nil && !kerr.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		}

		err = f.DeleteElasticsearchVersion(elasticsearchVersion.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {

				var shouldRunSuccessfully = func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					// Create Elasticsearch
					createAndWaitForRunning()

					By("Check for Elastic client")
					f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

					elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Creating new indices")
					err = elasticClient.CreateIndex(2)
					Expect(err).NotTo(HaveOccurred())

					By("Checking new indices")
					f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

					elasticClient.Stop()
					f.Tunnel.Close()

					By("Delete elasticsearch")
					err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for elasticsearch to be paused")
					f.EventuallyDormantDatabaseStatus(elasticsearch.ObjectMeta).Should(matcher.HavePaused())

					// Create Elasticsearch object again to resume it
					By("Create Elasticsearch: " + elasticsearch.Name)
					err = f.CreateElasticsearch(elasticsearch)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Wait for Running elasticsearch")
					f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())

					By("Check for Elastic client")
					f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

					elasticClient, err = f.GetElasticClient(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking new indices")
					f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

					elasticClient.Stop()
					f.Tunnel.Close()
				}

				Context("with Default Resource", func() {
					It("should run successfully", shouldRunSuccessfully)
				})

				Context("Custom Resource", func() {
					BeforeEach(func() {
						elasticsearch.Spec.PodTemplate.Spec.Resources = core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceMemory: resource.MustParse("512Mi"),
							},
						}
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should take Snapshot successfully", shouldRunSuccessfully)
				})

				Context("Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
					})

					Context("with Default Resource", func() {

						It("should run successfully", shouldRunSuccessfully)
					})

					Context("Custom Resource", func() {
						BeforeEach(func() {
							elasticsearch.Spec.Topology.Client.Resources = core.ResourceRequirements{
								Requests: core.ResourceList{
									core.ResourceMemory: resource.MustParse("128Mi"),
								},
							}
							elasticsearch.Spec.Topology.Master.Resources = core.ResourceRequirements{
								Requests: core.ResourceList{
									core.ResourceMemory: resource.MustParse("128Mi"),
								},
							}
							elasticsearch.Spec.Topology.Data.Resources = core.ResourceRequirements{
								Requests: core.ResourceList{
									core.ResourceMemory: resource.MustParse("128Mi"),
								},
							}
						})

						It("should run successfully", shouldRunSuccessfully)
					})

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should take Snapshot successfully", shouldRunSuccessfully)
					})

				})
			})

		})

		Context("Snapshot", func() {
			BeforeEach(func() {
				skipSnapshotDataChecking = false
				snapshot.Spec.DatabaseName = elasticsearch.Name
			})

			var shouldTakeSnapshot = func() {
				// Create and wait for running Elasticsearch
				createAndWaitForRunning()

				By("Create Secret")
				err := f.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())

				By("Create Snapshot")
				err = f.CreateSnapshot(snapshot)
				Expect(err).NotTo(HaveOccurred())

				By("Check for succeeded snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

				if !skipSnapshotDataChecking {
					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
				}
			}

			Context("In Local", func() {
				BeforeEach(func() {
					skipSnapshotDataChecking = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Local = &store.LocalSpec{
						MountPath: "/repo",
						VolumeSource: core.VolumeSource{
							EmptyDir: &core.EmptyDirVolumeSource{},
						},
					}
				})

				Context("With EmptyDir as Snapshot's backend", func() {

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("With PVC as Snapshot's backend", func() {

					BeforeEach(func() {
						snapshotPVC = f.GetPersistentVolumeClaim()
						By("Creating PVC for local backend snapshot")
						err := f.CreatePersistentVolumeClaim(snapshotPVC)
						Expect(err).NotTo(HaveOccurred())

						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
									ClaimName: snapshotPVC.Name,
								},
							},
						}
					})

					It("should delete Snapshot successfully", func() {
						shouldTakeSnapshot()

						By("Deleting Snapshot")
						f.DeleteSnapshot(snapshot.ObjectMeta)

						By("Waiting Snapshot to be deleted")
						f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())
					})
				})

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should take Snapshot successfully", shouldTakeSnapshot)
					})
				})
			})

			Context("In S3", func() {
				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &store.S3Spec{
						Bucket: os.Getenv(S3_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})
					It("should take Snapshot successfully", shouldTakeSnapshot)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should take Snapshot successfully", shouldTakeSnapshot)
					})
				})
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Azure", func() {
				BeforeEach(func() {
					secret = f.SecretForAzureBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Azure = &store.AzureSpec{
						Container: os.Getenv(AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &store.SwiftSpec{
						Container: os.Getenv(SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("Delete One Snapshot keeping others", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}

				})

				It("Delete One Snapshot keeping others", func() {
					// Create and wait for running elasticsearch
					createAndWaitForRunning()

					By("Create Secret")
					err := f.CreateSecret(secret)
					Expect(err).NotTo(HaveOccurred())

					By("Check for Elastic client")
					f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

					elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Creating new indices")
					err = elasticClient.CreateIndex(2)
					Expect(err).NotTo(HaveOccurred())

					By("Checking new indices")
					f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

					elasticClient.Stop()
					f.Tunnel.Close()

					By("Create Snapshot")
					err = f.CreateSnapshot(snapshot)
					Expect(err).NotTo(HaveOccurred())

					By("Check for Succeeded snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					if !skipSnapshotDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					oldSnapshot := snapshot

					// create new Snapshot
					snapshot := f.Snapshot()
					snapshot.Spec.DatabaseName = elasticsearch.Name
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}

					By("Create Snapshot")
					err = f.CreateSnapshot(snapshot)
					Expect(err).NotTo(HaveOccurred())

					By("Check for Succeeded snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					if !skipSnapshotDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					By(fmt.Sprintf("Delete Snapshot %v", snapshot.Name))
					err = f.DeleteSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Deleting Snapshot")
					f.EventuallySnapshot(elasticsearch.ObjectMeta).Should(BeFalse())

					if !skipSnapshotDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
					}

					snapshot = oldSnapshot

					By(fmt.Sprintf("Old Snapshot %v Still Exists", snapshot.Name))
					_, err = f.GetSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					if !skipSnapshotDataChecking {
						By(fmt.Sprintf("Check for old snapshot %v data", snapshot.Name))
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}
				})
			})
		})

		Context("Initialize", func() {

			BeforeEach(func() {
				skipSnapshotDataChecking = false
				secret = f.SecretForS3Backend()
				snapshot.Spec.StorageSecretName = secret.Name
				snapshot.Spec.S3 = &store.S3Spec{
					Bucket: os.Getenv(S3_BUCKET_NAME),
				}
				snapshot.Spec.DatabaseName = elasticsearch.Name
			})

			var shouldInitialize = func() {
				// Create and wait for running Elasticsearch
				createAndWaitForRunning()

				By("Create Secret")
				f.CreateSecret(secret)

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Creating new indices")
				err = elasticClient.CreateIndex(2)
				Expect(err).NotTo(HaveOccurred())

				By("Checking new indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()

				By("Create Snapshot")
				f.CreateSnapshot(snapshot)

				By("Check for succeeded snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

				if !skipSnapshotDataChecking {
					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
				}

				oldElasticsearch, err := f.GetElasticsearch(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				garbageElasticsearch.Items = append(garbageElasticsearch.Items, *oldElasticsearch)

				By("Create elasticsearch from snapshot")
				*elasticsearch = *f.CombinedElasticsearch()
				elasticsearch.Spec.Init = &api.InitSpec{
					SnapshotSource: &api.SnapshotSourceSpec{
						Namespace: snapshot.Namespace,
						Name:      snapshot.Name,
					},
				}

				// Create and wait for running Elasticsearch
				createAndWaitForRunning()

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err = f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Checking indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()
			}

			Context("-", func() {
				It("should initialize database successfully", shouldInitialize)
			})

			Context("with local volume", func() {

				BeforeEach(func() {
					snapshotPVC = f.GetPersistentVolumeClaim()
					By("Creating PVC for local backend snapshot")
					err := f.CreatePersistentVolumeClaim(snapshotPVC)
					Expect(err).NotTo(HaveOccurred())

					skipSnapshotDataChecking = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Backend = store.Backend{
						Local: &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
									ClaimName: snapshotPVC.Name,
								},
							},
						},
					}
				})

				It("should initialize database successfully", shouldInitialize)

			})

			Context("with SSL disabled", func() {
				BeforeEach(func() {
					elasticsearch.Spec.EnableSSL = false
				})

				It("should initialize database successfully", shouldInitialize)
			})

			Context("with Dedicated elasticsearch", func() {
				BeforeEach(func() {
					elasticsearch = f.DedicatedElasticsearch()
					snapshot.Spec.DatabaseName = elasticsearch.Name
				})
				It("should initialize database successfully", shouldInitialize)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should initialize database successfully", shouldInitialize)
				})
			})
		})

		Context("Resume", func() {
			var usedInitialized bool
			BeforeEach(func() {
				usedInitialized = false
			})

			var shouldResumeSuccessfully = func() {
				// Create and wait for running Elasticsearch
				createAndWaitForRunning()

				By("Delete elasticsearch")
				f.DeleteElasticsearch(elasticsearch.ObjectMeta)

				By("Wait for elasticsearch to be paused")
				f.EventuallyDormantDatabaseStatus(elasticsearch.ObjectMeta).Should(matcher.HavePaused())

				// Create Elasticsearch object again to resume it
				By("Create Elasticsearch: " + elasticsearch.Name)
				err = f.CreateElasticsearch(elasticsearch)
				Expect(err).NotTo(HaveOccurred())

				By("Wait for DormantDatabase to be deleted")
				f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

				By("Wait for Running elasticsearch")
				f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())

				es, err := f.GetElasticsearch(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				*elasticsearch = *es
				if usedInitialized {
					_, ok := elasticsearch.Annotations[api.AnnotationInitialized]
					Expect(ok).Should(BeTrue())
				}
			}

			Context("-", func() {
				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("with SSL disabled", func() {
				BeforeEach(func() {
					elasticsearch.Spec.EnableSSL = false
				})

				It("should initialize database successfully", shouldResumeSuccessfully)
			})

			Context("with Dedicated elasticsearch", func() {
				BeforeEach(func() {
					elasticsearch = f.DedicatedElasticsearch()
					snapshot.Spec.DatabaseName = elasticsearch.Name
				})
				It("should initialize database successfully", shouldResumeSuccessfully)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should initialize database successfully", shouldResumeSuccessfully)
				})
			})
		})

		Context("SnapshotScheduler", func() {
			AfterEach(func() {
				f.DeleteSecret(secret.ObjectMeta)
			})

			BeforeEach(func() {
				secret = f.SecretForLocalBackend()
			})

			Context("With Startup", func() {
				BeforeEach(func() {
					elasticsearch.Spec.BackupSchedule = &api.BackupScheduleSpec{
						CronExpression: "@every 1m",
						Backend: store.Backend{
							StorageSecretName: secret.Name,
							Local: &store.LocalSpec{
								MountPath: "/repo",
								VolumeSource: core.VolumeSource{
									EmptyDir: &core.EmptyDirVolumeSource{},
								},
							},
						},
					}
				})

				var shouldStartupSchedular = func() {
					// Create and wait for running Elasticsearch
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Count multiple Snapshot")
					f.EventuallySnapshotCount(elasticsearch.ObjectMeta).Should(matcher.MoreThan(3))
				}

				Context("-", func() {
					It("should run schedular successfully", shouldStartupSchedular)
				})

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should run schedular successfully", shouldStartupSchedular)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						elasticsearch.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 1m",
							Backend: store.Backend{
								StorageSecretName: secret.Name,
								Local: &store.LocalSpec{
									MountPath: "/repo",
									VolumeSource: core.VolumeSource{
										EmptyDir: &core.EmptyDirVolumeSource{},
									},
								},
							},
						}
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})
					It("should run schedular successfully", shouldStartupSchedular)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should run schedular successfully", shouldStartupSchedular)
					})
				})
			})

			Context("With Update", func() {
				var shouldScheduleWithUpdate = func() {
					// Create and wait for running Elasticsearch
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update elasticsearch")
					_, err = f.TryPatchElasticsearch(elasticsearch.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
						in.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 1m",
							Backend: store.Backend{
								StorageSecretName: secret.Name,
								Local: &store.LocalSpec{
									MountPath: "/repo",
									VolumeSource: core.VolumeSource{
										EmptyDir: &core.EmptyDirVolumeSource{},
									},
								},
							},
						}

						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Count multiple Snapshot")
					f.EventuallySnapshotCount(elasticsearch.ObjectMeta).Should(matcher.MoreThan(3))
				}
				Context("-", func() {
					It("should run schedular successfully", shouldScheduleWithUpdate)
				})
			})
		})

		Context("Termination Policy", func() {
			BeforeEach(func() {
				skipSnapshotDataChecking = false
				secret = f.SecretForS3Backend()
				snapshot.Spec.StorageSecretName = secret.Name
				snapshot.Spec.S3 = &store.S3Spec{
					Bucket: os.Getenv(S3_BUCKET_NAME),
				}
				snapshot.Spec.DatabaseName = elasticsearch.Name
			})

			AfterEach(func() {
				if snapshot != nil {
					By("Delete Existing snapshot")
					err := f.DeleteSnapshot(snapshot.ObjectMeta)
					if err != nil {
						if kerr.IsNotFound(err) {
							// Elasticsearch was not created. Hence, rest of cleanup is not necessary.
							return
						}
						Expect(err).NotTo(HaveOccurred())
					}
				}
			})

			var shouldRunWithSnapshot = func() {
				// Create and wait for running Elasticsearch
				createAndWaitForRunning()

				By("Create Secret")
				f.CreateSecret(secret)

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Creating new indices")
				err = elasticClient.CreateIndex(2)
				Expect(err).NotTo(HaveOccurred())

				By("Checking new indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()

				By("Create Snapshot")
				f.CreateSnapshot(snapshot)

				By("Check for succeeded snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

				if !skipSnapshotDataChecking {
					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
				}
			}

			Context("with TerminationPolicyDoNotTerminate", func() {

				BeforeEach(func() {
					skipSnapshotDataChecking = true
					elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running Elasticsearch
					createAndWaitForRunning()

					By("Delete elasticsearch")
					err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("Elasticsearch is not paused. Check for elasticsearch")
					f.EventuallyElasticsearch(elasticsearch.ObjectMeta).Should(BeTrue())

					By("Check for Running elasticsearch")
					f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())

					By("Update elasticsearch to set spec.terminationPolicy = Pause")
					f.TryPatchElasticsearch(elasticsearch.ObjectMeta, func(in *api.Elasticsearch) *api.Elasticsearch {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
				})
			})

			Context("with TerminationPolicyPause (default)", func() {
				var shouldRunWithTerminationPause = func() {
					shouldRunWithSnapshot()

					By("Delete elasticsearch")
					err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means elasticsearch object is deleted
					By("Wait for elasticsearch to be paused")
					f.EventuallyDormantDatabaseStatus(elasticsearch.ObjectMeta).Should(matcher.HavePaused())

					By("Check for intact snapshot")
					_, err := f.GetSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					if !skipSnapshotDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					// Create Elasticsearch object again to resume it
					By("Create (pause) Elasticsearch: " + elasticsearch.Name)
					err = f.CreateElasticsearch(elasticsearch)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Wait for Running elasticsearch")
					f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())

					By("Check for Elastic client")
					f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

					elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking new indices")
					f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

					elasticClient.Stop()
					f.Tunnel.Close()
				}

				It("should create dormantdatabase successfully", shouldRunWithTerminationPause)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should create dormantdatabase successfully", shouldRunWithTerminationPause)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})
					It("should initialize database successfully", shouldRunWithTerminationPause)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should initialize database successfully", shouldRunWithTerminationPause)
					})
				})
			})

			Context("with TerminationPolicyDelete", func() {
				BeforeEach(func() {
					elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				var shouldRunWithTerminationDelete = func() {
					shouldRunWithSnapshot()

					By("Delete elasticsearch")
					err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until elasticsearch is deleted")
					f.EventuallyElasticsearch(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Check for deleted PVCs")
					f.EventuallyPVCCount(elasticsearch.ObjectMeta).Should(Equal(0))

					By("Check for intact Secrets")
					f.EventuallyDBSecretCount(elasticsearch.ObjectMeta).ShouldNot(Equal(0))

					By("Check for intact snapshot")
					_, err := f.GetSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					if !skipSnapshotDataChecking {
						By("Check for intact snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					By("Delete snapshot")
					err = f.DeleteSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					if !skipSnapshotDataChecking {
						By("Check for deleted snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
					}
				}

				It("should run with TerminationPolicyDelete", shouldRunWithTerminationDelete)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})
					It("should run with TerminationPolicyDelete", shouldRunWithTerminationDelete)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyDelete
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})
					It("should initialize database successfully", shouldRunWithTerminationDelete)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should initialize database successfully", shouldRunWithTerminationDelete)
					})
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				var shouldRunWithTerminationWipeOut = func() {
					shouldRunWithSnapshot()

					By("Delete elasticsearch")
					err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until elasticsearch is deleted")
					f.EventuallyElasticsearch(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

					By("Check for deleted PVCs")
					f.EventuallyPVCCount(elasticsearch.ObjectMeta).Should(Equal(0))

					By("Check for deleted Secrets")
					f.EventuallyDBSecretCount(elasticsearch.ObjectMeta).Should(Equal(0))

					By("Check for deleted Snapshots")
					f.EventuallySnapshotCount(snapshot.ObjectMeta).Should(Equal(0))

					if !skipSnapshotDataChecking {
						By("Check for deleted snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
					}
				}

				It("should run with TerminationPolicyWipeOut", shouldRunWithTerminationWipeOut)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})
					It("should run with TerminationPolicyDelete", shouldRunWithTerminationWipeOut)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						snapshot.Spec.DatabaseName = elasticsearch.Name
						elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})
					It("should initialize database successfully", shouldRunWithTerminationWipeOut)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should initialize database successfully", shouldRunWithTerminationWipeOut)
					})
				})
			})
		})

		Context("Environment Variables", func() {

			allowedEnvList := []core.EnvVar{
				{
					Name:  "CLUSTER_NAME",
					Value: "kubedb-es-e2e-cluster",
				},
				{
					Name:  "NUMBER_OF_MASTERS",
					Value: "1",
				},
				{
					Name:  "ES_JAVA_OPTS",
					Value: "-Xms256m -Xmx256m",
				},
				{
					Name:  "REPO_LOCATIONS",
					Value: "/backup",
				},
				{
					Name:  "MEMORY_LOCK",
					Value: "true",
				},
				{
					Name:  "HTTP_ENABLE",
					Value: "true",
				},
			}

			forbiddenEnvList := []core.EnvVar{
				{
					Name:  "NODE_NAME",
					Value: "kubedb-es-e2e-node",
				},
				{
					Name:  "NODE_MASTER",
					Value: "true",
				},
				{
					Name:  "NODE_DATA",
					Value: "true",
				},
			}

			var shouldRunSuccessfully = func() {
				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create Elasticsearch
				createAndWaitForRunning()

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Creating new indices")
				err = elasticClient.CreateIndex(2)
				Expect(err).NotTo(HaveOccurred())

				By("Checking new indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()

				By("Delete elasticsearch")
				err = f.DeleteElasticsearch(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Wait for elasticsearch to be paused")
				f.EventuallyDormantDatabaseStatus(elasticsearch.ObjectMeta).Should(matcher.HavePaused())

				// Create Elasticsearch object again to resume it
				By("Create Elasticsearch: " + elasticsearch.Name)
				err = f.CreateElasticsearch(elasticsearch)
				Expect(err).NotTo(HaveOccurred())

				By("Wait for DormantDatabase to be deleted")
				f.EventuallyDormantDatabase(elasticsearch.ObjectMeta).Should(BeFalse())

				By("Wait for Running elasticsearch")
				f.EventuallyElasticsearchRunning(elasticsearch.ObjectMeta).Should(BeTrue())

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err = f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Checking new indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()
			}

			Context("With allowed Envs", func() {

				var shouldRunWithAllowedEnvs = func() {
					elasticsearch.Spec.PodTemplate.Spec.Env = allowedEnvList
					shouldRunSuccessfully()

					podName := f.GetClientPodName(elasticsearch)

					By("Checking pod started with given envs")
					pod, err := f.KubeClient().CoreV1().Pods(elasticsearch.Namespace).Get(podName, metav1.GetOptions{})
					Expect(err).NotTo(HaveOccurred())

					out, err := exec_util.ExecIntoPod(f.RestConfig(), pod, "env")
					Expect(err).NotTo(HaveOccurred())
					for _, env := range allowedEnvList {
						Expect(out).Should(ContainSubstring(env.Name + "=" + env.Value))
					}
				}

				Context("-", func() {
					It("should run successfully with given envs", shouldRunWithAllowedEnvs)
				})

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should run successfully with given envs", shouldRunWithAllowedEnvs)
				})

				Context("with Dedicated elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						snapshot.Spec.DatabaseName = elasticsearch.Name
					})
					It("should run successfully with given envs", shouldRunWithAllowedEnvs)

					Context("with SSL disabled", func() {
						BeforeEach(func() {
							elasticsearch.Spec.EnableSSL = false
						})

						It("should run successfully with given envs", shouldRunWithAllowedEnvs)
					})
				})
			})

			Context("With forbidden Envs", func() {

				It("should reject to create Elasticsearch CRD", func() {
					for _, env := range forbiddenEnvList {
						elasticsearch.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							env,
						}

						By("Creating Elasticsearch with " + env.Name + " env var.")
						err := f.CreateElasticsearch(elasticsearch)
						Expect(err).To(HaveOccurred())
					}
				})
			})

			Context("Update Envs", func() {

				It("should reject to update Envs", func() {
					elasticsearch.Spec.PodTemplate.Spec.Env = allowedEnvList

					shouldRunSuccessfully()

					By("Updating Envs")
					_, _, err := util.PatchElasticsearch(f.ExtClient().KubedbV1alpha1(), elasticsearch, func(in *api.Elasticsearch) *api.Elasticsearch {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  "CLUSTER_NAME",
								Value: "kubedb-es-e2e-cluster-patched",
							},
						}
						return in
					})
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("Custom Configuration", func() {

			var userConfig *core.ConfigMap

			var shouldRunWithCustomConfig = func() {
				userConfig.Data = map[string]string{
					"common-config.yaml": f.GetCommonConfig(),
					"master-config.yaml": f.GetMasterConfig(),
					"client-config.yaml": f.GetClientConfig(),
					"data-config.yaml":   f.GetDataConfig(),
				}

				By("Creating configMap: " + userConfig.Name)
				err := f.CreateConfigMap(userConfig)
				Expect(err).NotTo(HaveOccurred())

				elasticsearch.Spec.ConfigSource = &core.VolumeSource{
					ConfigMap: &core.ConfigMapVolumeSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: userConfig.Name,
						},
					},
				}

				// Create Elasticsearch
				createAndWaitForRunning()

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Reading Nodes information")
				settings, err := elasticClient.GetAllNodesInfo()
				Expect(err).NotTo(HaveOccurred())

				By("Checking nodes are using provided config")
				Expect(f.IsUsingProvidedConfig(settings)).Should(BeTrue())

				elasticClient.Stop()
				f.Tunnel.Close()
			}

			Context("With Topology", func() {
				BeforeEach(func() {
					elasticsearch = f.DedicatedElasticsearch()
					userConfig = f.GetCustomConfig()
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					f.DeleteConfigMap(userConfig.ObjectMeta)
				})

				It("should use config provided in config files", shouldRunWithCustomConfig)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should run successfully with given envs", shouldRunWithCustomConfig)
				})
			})

			Context("Without Topology", func() {
				BeforeEach(func() {
					userConfig = f.GetCustomConfig()
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					f.DeleteConfigMap(userConfig.ObjectMeta)
				})

				It("should use config provided in config files", shouldRunWithCustomConfig)

				Context("with SSL disabled", func() {
					BeforeEach(func() {
						elasticsearch.Spec.EnableSSL = false
					})

					It("should run successfully with given envs", shouldRunWithCustomConfig)
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create Elasticsearch
				createAndWaitForRunning()

				By("Check for Elastic client")
				f.EventuallyElasticsearchClientReady(elasticsearch.ObjectMeta).Should(BeTrue())

				elasticClient, err := f.GetElasticClient(elasticsearch.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Creating new indices")
				err = elasticClient.CreateIndex(2)
				Expect(err).NotTo(HaveOccurred())

				By("Checking new indices")
				f.EventuallyElasticsearchIndicesCount(elasticClient).Should(Equal(3))

				elasticClient.Stop()
				f.Tunnel.Close()
			}

			Context("Ephemeral", func() {

				Context("Combined Elasticsearch", func() {

					BeforeEach(func() {
						elasticsearch.Spec.StorageType = api.StorageTypeEphemeral
						elasticsearch.Spec.Storage = nil
						elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("Dedicated Elasticsearch", func() {
					BeforeEach(func() {
						elasticsearch = f.DedicatedElasticsearch()
						elasticsearch.Spec.StorageType = api.StorageTypeEphemeral
						elasticsearch.Spec.Topology.Master.Storage = nil
						elasticsearch.Spec.Topology.Client.Storage = nil
						elasticsearch.Spec.Topology.Data.Storage = nil
						elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						elasticsearch.Spec.StorageType = api.StorageTypeEphemeral
						elasticsearch.Spec.Storage = nil
						elasticsearch.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create Elasticsearch object", func() {

						By("Creating Elasticsearch: " + elasticsearch.Name)
						err := f.CreateElasticsearch(elasticsearch)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
