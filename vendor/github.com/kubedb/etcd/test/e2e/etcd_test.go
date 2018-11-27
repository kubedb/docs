package e2e_test

import (
	"fmt"
	"os"

	"github.com/appscode/go/log"
	meta_util "github.com/appscode/kutil/meta"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/etcd/test/e2e/framework"
	"github.com/kubedb/etcd/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	store "kmodules.xyz/objectstore-api/api/v1"
)

var _ = Describe("Etcd", func() {
	var (
		err           error
		f             *framework.Invocation
		etcd          *api.Etcd
		garbageEtcd   *api.EtcdList
		etcdVersion   *catalog.EtcdVersion
		snapshot      *api.Snapshot
		secret        *core.Secret
		skipMessage   string
		skipDataCheck bool
	)

	BeforeEach(func() {
		f = root.Invoke()
		etcd = f.Etcd()
		garbageEtcd = new(api.EtcdList)
		snapshot = f.Snapshot()
		etcdVersion = f.EtcdVersion()
		skipMessage = ""
	})

	var createAndWaitForRunning = func() {
		By("Ensuring EtcdVersion crd")
		err := f.CreateEtcdVersion(etcdVersion)
		Expect(err).NotTo(HaveOccurred())

		By("Create Etcd: " + etcd.Name)
		err = f.CreateEtcd(etcd)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running etcd")
		f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(etcd.ObjectMeta).Should(BeTrue())
	}

	var deleteTestResource = func() {
		if etcd == nil {
			log.Infoln("Skipping cleanup. Reason: etcd is nil.")
			return
		}

		By("Check if etcd " + etcd.Name + " exists.")
		et, err := f.GetEtcd(etcd.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Etcd was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete etcd")
		err = f.DeleteEtcd(etcd.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infof("Skipping rest of the cleanup. Reason: Etcd  %s not found", etcd.Name)
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if et.Spec.TerminationPolicy == api.TerminationPolicyPause {

			By("Wait for etcd to be paused")
			f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

			By("Set DormantDatabase Spec.WipeOut to true")
			_, err := f.PatchDormantDatabase(etcd.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(etcd.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for etcd resources to be wipedOut")
		f.EventuallyWipedOut(etcd.ObjectMeta).Should(Succeed())
	}

	AfterEach(func() {
		// Delete test resource
		deleteTestResource()

		for _, et := range garbageEtcd.Items {
			*etcd = et
			// Delete test resource
			deleteTestResource()
		}

		if !skipDataCheck {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
		}

		if secret != nil {
			f.DeleteSecret(secret.ObjectMeta)
		}

		err = f.DeleteEtcdVersion(etcdVersion.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Test", func() {
		BeforeEach(func() {
			if f.StorageClass == "" {
				Skip("Missing StorageClassName. Provide as flag to test this.")
			}
			etcd.Spec.Storage = f.EtcdPVCSpec()
		})

		Context("General", func() {

			Context("Without PVC", func() {
				BeforeEach(func() {
					etcd.Spec.Storage = nil
				})
				It("should run successfully", createAndWaitForRunning)
			})

			Context("With PVC", func() {
				It("should run successfully", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}
					// Create Etcd
					createAndWaitForRunning()

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())
				})
			})
		})

		Context("DoNotTerminate", func() {
			BeforeEach(func() {
				etcd.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
			})

			It("should work successfully", func() {
				// Create and wait for running Etcd
				createAndWaitForRunning()

				By("Delete etcd")
				err = f.DeleteEtcd(etcd.ObjectMeta)
				Expect(err).Should(HaveOccurred())

				By("Etcd is not paused. Check for etcd")
				f.EventuallyEtcd(etcd.ObjectMeta).Should(BeTrue())

				By("Check for Running etcd")
				f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

				By("Update etcd to set spec.terminationPolicy = Pause")
				f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
					in.Spec.TerminationPolicy = api.TerminationPolicyPause
					return in
				})
			})
		})

		Context("Snapshot", func() {

			AfterEach(func() {
				f.DeleteSecret(secret.ObjectMeta)
			})

			BeforeEach(func() {
				skipDataCheck = false
				snapshot.Spec.DatabaseName = etcd.Name
			})

			var shouldTakeSnapshot = func() {
				// Create and wait for running Etcd
				createAndWaitForRunning()

				By("Create Secret")
				err := f.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())

				By("Create Snapshot")
				err = f.CreateSnapshot(snapshot)
				Expect(err).NotTo(HaveOccurred())

				By("Check for Succeeded snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
			}

			Context("In Local", func() {
				BeforeEach(func() {
					skipDataCheck = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Local = f.LocalStorageSpec()
				})

				Context("With EmptyDir as Snapshot's backend", func() {

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("With PVC as Snapshot's backend", func() {
					var snapPVC *core.PersistentVolumeClaim

					BeforeEach(func() {
						snapPVC = f.GetPersistentVolumeClaim()
						err := f.CreatePersistentVolumeClaim(snapPVC)
						Expect(err).NotTo(HaveOccurred())

						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
									ClaimName: snapPVC.Name,
								},
							},
						}
					})

					AfterEach(func() {
						f.DeletePersistentVolumeClaim(snapPVC.ObjectMeta)
					})

					It("should delete Snapshot successfully", func() {
						shouldTakeSnapshot()

						By("Deleting Snapshot")
						f.DeleteSnapshot(snapshot.ObjectMeta)

						By("Waiting Snapshot to be deleted")
						f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())
					})
				})

				// Additional
				Context("Without PVC", func() {
					BeforeEach(func() {
						etcd.Spec.Storage = nil
					})
					It("should run successfully", shouldTakeSnapshot)
				})
			})

			Context("In S3", func() {
				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &store.S3Spec{
						Bucket: os.Getenv(framework.S3_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(framework.GCS_BUCKET_NAME),
					}
				})

				Context("Without Init", func() {
					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("With Init", func() {
					BeforeEach(func() {
						etcd.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/etcd-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("Delete One Snapshot keeping others", func() {
					BeforeEach(func() {
						etcd.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/etcd-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create and wait for running Etcd
						createAndWaitForRunning()

						By("Create Secret")
						err := f.CreateSecret(secret)
						Expect(err).NotTo(HaveOccurred())

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						oldSnapshot := snapshot

						// create new Snapshot
						snapshot := f.Snapshot()
						snapshot.Spec.DatabaseName = etcd.Name
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.GCS = &store.GCSSpec{
							Bucket: os.Getenv(framework.GCS_BUCKET_NAME),
						}

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						By(fmt.Sprintf("Delete Snapshot %v", snapshot.Name))
						err = f.DeleteSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for Deleting Snapshot")
						f.EventuallySnapshot(etcd.ObjectMeta).Should(BeFalse())
						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
						}

						snapshot = oldSnapshot

						By(fmt.Sprintf("Old Snapshot %v Still Exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataCheck {
							By(fmt.Sprintf("Check for old snapshot %v data", snapshot.Name))
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}
					})
				})

			})

			Context("In Azure", func() {
				BeforeEach(func() {
					secret = f.SecretForAzureBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Azure = &store.AzureSpec{
						Container: os.Getenv(framework.AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &store.SwiftSpec{
						Container: os.Getenv(framework.SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})
		})

		Context("Initialize", func() {
			Context("With Script", func() {
				BeforeEach(func() {
					etcd.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/etcd-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", func() {
					// Create Postgres
					createAndWaitForRunning()

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())
				})

			})

			Context("With Snapshot", func() {
				AfterEach(func() {
					f.DeleteSecret(secret.ObjectMeta)
				})

				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(framework.GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = etcd.Name
				})

				It("should run successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Create Secret")
					f.CreateSecret(secret)

					By("Create Snapshot")
					f.CreateSnapshot(snapshot)

					By("Check for Succeeded snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())

					oldEtcd, err := f.GetEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// insert oldEtcd into garbage list so that it's resources can be cleaned after test
					garbageEtcd.Items = append(garbageEtcd.Items, *oldEtcd)

					By("Create etcd from snapshot")
					etcd = f.Etcd()
					if f.StorageClass != "" {
						etcd.Spec.Storage = f.EtcdPVCSpec()
					}
					etcd.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())
				})
			})
		})

		Context("Resume", func() {

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					// Delete without caring if DB is resumed
					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					_, err = f.GetEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())
				})
			})

			Context("with init Script", func() {
				BeforeEach(func() {
					etcd.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/etcd-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					etcd, err := f.GetEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking etcd does not have kubedb.com/initialized annotation")
					_, err = meta_util.GetString(etcd.Annotations, api.AnnotationInitialized)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("With Snapshot Init", func() {
				AfterEach(func() {
					f.DeleteSecret(secret.ObjectMeta)
				})
				BeforeEach(func() {
					skipDataCheck = false
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(framework.GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = etcd.Name
				})
				It("should resume successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Create Secret")
					f.CreateSecret(secret)

					By("Create Snapshot")
					f.CreateSnapshot(snapshot)

					By("Check for Succeeded snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())

					oldEtcd, err := f.GetEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// insert oldEtcd into garbage list so that it's resources can be cleaned after test
					garbageEtcd.Items = append(garbageEtcd.Items, *oldEtcd)

					By("Create etcd from snapshot")
					etcd = f.Etcd()
					if f.StorageClass != "" {
						etcd.Spec.Storage = f.EtcdPVCSpec()
					}

					etcd.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					etcd, err = f.GetEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Etcd has kubedb.com/initialized annotation")
					_, err = meta_util.GetString(etcd.Annotations, api.AnnotationInitialized)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Multiple times with init script", func() {
				BeforeEach(func() {
					etcd.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/etcd-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")
						By("Delete etcd")
						err = f.DeleteEtcd(etcd.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for etcd to be paused")
						f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

						// Create Etcd object again to resume it
						By("Create Etcd: " + etcd.Name)
						err = f.CreateEtcd(etcd)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

						By("Wait for Running etcd")
						f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

						_, err := f.GetEtcd(etcd.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Checking Inserted Document")
						f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

						By("Checking etcd does not have kubedb.com/initialized annotation")
						_, err = meta_util.GetString(etcd.Annotations, api.AnnotationInitialized)
						Expect(err).To(HaveOccurred())
					}
				})
			})

		})

		Context("SnapshotScheduler", func() {
			AfterEach(func() {
				f.DeleteSecret(secret.ObjectMeta)
			})

			Context("With Startup", func() {

				var shouldStartupSchedular = func() {
					By("Create Secret")
					f.CreateSecret(secret)

					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(etcd.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from Etcd")
					_, err = f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(etcd.ObjectMeta).Should(Succeed())
				}

				Context("with local", func() {
					BeforeEach(func() {
						secret = f.SecretForLocalBackend()
						etcd.Spec.BackupSchedule = f.LocalBackupScheduleSpec(secret.Name)
					})

					It("should run scheduler successfully", shouldStartupSchedular)
				})

				Context("with GCS and PVC", func() {
					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						etcd.Spec.BackupSchedule = f.GCSBackupScheduleSpec(secret.Name)
					})

					It("should run scheduler successfully", shouldStartupSchedular)
				})
			})

			Context("With Update - with Local", func() {
				BeforeEach(func() {
					secret = f.SecretForLocalBackend()
				})
				It("should run scheduler successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update etcd")
					_, err = f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
						in.Spec.BackupSchedule = f.LocalBackupScheduleSpec(secret.Name)
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(etcd.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from Etcd")
					_, err = f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(etcd.ObjectMeta).Should(Succeed())
				})
			})

			Context("Re-Use DormantDatabase's scheduler", func() {
				BeforeEach(func() {
					secret = f.SecretForLocalBackend()
				})
				It("should re-use scheduler successfully", func() {
					// Create and wait for running Etcd
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update etcd")
					_, err = f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
						in.Spec.BackupSchedule = f.LocalBackupScheduleSpec(secret.Name)
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Insert Document Inside DB")
					f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(etcd.ObjectMeta).Should(matcher.MoreThan(3))

					By("Delete etcd")
					err = f.DeleteEtcd(etcd.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for etcd to be paused")
					f.EventuallyDormantDatabaseStatus(etcd.ObjectMeta).Should(matcher.HavePaused())

					// Create Etcd object again to resume it
					By("Create Etcd: " + etcd.Name)
					err = f.CreateEtcd(etcd)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(etcd.ObjectMeta).Should(BeFalse())

					By("Wait for Running etcd")
					f.EventuallyEtcdRunning(etcd.ObjectMeta).Should(BeTrue())

					By("Checking Inserted Document")
					f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(etcd.ObjectMeta).Should(matcher.MoreThan(5))

					By("Remove Backup Scheduler from Etcd")
					_, err = f.PatchEtcd(etcd.ObjectMeta, func(in *api.Etcd) *api.Etcd {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(etcd.ObjectMeta).Should(Succeed())
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}
				// Create Etcd
				createAndWaitForRunning()

				By("Insert Key into DB")
				f.EventuallySetKey(etcd.ObjectMeta).Should(BeTrue())

				By("Checking Key Exist")
				f.EventuallyKeyExists(etcd.ObjectMeta).Should(BeTrue())
			}

			Context("Ephemeral", func() {

				Context("General Behaviour", func() {

					BeforeEach(func() {
						skipDataCheck = true
						etcd.Spec.StorageType = api.StorageTypeEphemeral
						etcd.Spec.Storage = nil
						etcd.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						etcd.Spec.StorageType = api.StorageTypeEphemeral
						etcd.Spec.Storage = nil
						etcd.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create Etcd object", func() {

						By("Creating Etcd: " + etcd.Name)
						err := f.CreateEtcd(etcd)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
