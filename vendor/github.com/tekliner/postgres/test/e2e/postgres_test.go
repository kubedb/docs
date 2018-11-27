package e2e_test

import (
	"fmt"
	"os"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/postgres/test/e2e/framework"
	"github.com/kubedb/postgres/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	store "kmodules.xyz/objectstore-api/api/v1"
)

const (
	S3_BUCKET_NAME          = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME         = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME    = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME    = "SWIFT_CONTAINER_NAME"
	POSTGRES_DB             = "POSTGRES_DB"
	POSTGRES_PASSWORD       = "POSTGRES_PASSWORD"
	PGDATA                  = "PGDATA"
	POSTGRES_USER           = "POSTGRES_USER"
	POSTGRES_INITDB_ARGS    = "POSTGRES_INITDB_ARGS"
	POSTGRES_INITDB_WALDIR  = "POSTGRES_INITDB_WALDIR"
	POSTGRES_INITDB_XLOGDIR = "POSTGRES_INITDB_XLOGDIR"
)

var _ = Describe("Postgres", func() {
	var (
		err                      error
		f                        *framework.Invocation
		postgres                 *api.Postgres
		garbagePostgres          *api.PostgresList
		postgresVersion          *catalog.PostgresVersion
		snapshot                 *api.Snapshot
		secret                   *core.Secret
		skipMessage              string
		skipSnapshotDataChecking bool
		skipWalDataChecking      bool
		dbName                   string
		dbUser                   string
	)

	BeforeEach(func() {
		f = root.Invoke()
		postgres = f.Postgres()
		postgresVersion = f.PostgresVersion()
		garbagePostgres = new(api.PostgresList)
		snapshot = f.Snapshot()
		secret = new(core.Secret)
		skipMessage = ""
		skipSnapshotDataChecking = true
		skipWalDataChecking = true
		dbName = "postgres"
		dbUser = "postgres"
	})

	var createAndWaitForRunning = func() {

		By("Ensuring PostgresVersion crd: " + postgresVersion.Spec.DB.Image)
		err = f.CreatePostgresVersion(postgresVersion)
		Expect(err).NotTo(HaveOccurred())

		By("Creating Postgres: " + postgres.Name)
		err = f.CreatePostgres(postgres)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running postgres")
		f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

		By("Waiting for database to be ready")
		f.EventuallyPingDatabase(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())
	}

	var testGeneralBehaviour = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}
		// Create Postgres
		createAndWaitForRunning()

		By("Creating Schema")
		f.EventuallyCreateSchema(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

		By("Creating Table")
		f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

		By("Checking Table")
		f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

		By("Delete postgres")
		err = f.DeletePostgres(postgres.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for postgres to be paused")
		f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

		// Create Postgres object again to resume it
		By("Create Postgres: " + postgres.Name)
		err = f.CreatePostgres(postgres)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for DormantDatabase to be deleted")
		f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

		By("Wait for Running postgres")
		f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

		By("Checking Table")
		f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))
	}

	var shouldTakeSnapshot = func() {
		// Create and wait for running Postgres
		createAndWaitForRunning()

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeeded snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipSnapshotDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var shouldInsertDataAndTakeSnapshot = func() {
		// Create and wait for running Postgres
		createAndWaitForRunning()

		By("Creating Schema")
		f.EventuallyCreateSchema(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

		By("Creating Table")
		f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

		By("Checking Table")
		f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

		By("Create Secret")
		err = f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeeded snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipSnapshotDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var deleteTestResource = func() {
		if postgres == nil {
			Skip("Skipping")
		}

		By("Check if Postgres " + postgres.Name + " exists.")
		pg, err := f.GetPostgres(postgres.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Postgres was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete postgres: " + postgres.Name)
		err = f.DeletePostgres(postgres.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Postgres was not created. Hence, rest of cleanup is not necessary.
				log.Infof("Skipping rest of cleanup. Reason: Postgres %s is not found.", postgres.Name)
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if pg.Spec.TerminationPolicy == api.TerminationPolicyPause {

			By("Wait for postgres to be paused")
			f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

			By("Set DormantDatabase Spec.WipeOut to true")
			_, err := f.PatchDormantDatabase(postgres.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(postgres.ObjectMeta)
			if !kerr.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}

		}

		By("Wait for postgres resources to be wipedOut")
		f.EventuallyWipedOut(postgres.ObjectMeta).Should(Succeed())

		if postgres.Spec.Archiver != nil && !skipWalDataChecking {
			By("Checking wal data has been removed")
			f.EventuallyWalDataFound(postgres).Should(BeFalse())
		}
	}

	AfterEach(func() {
		// Delete test resource
		deleteTestResource()

		for _, pg := range garbagePostgres.Items {
			*postgres = pg
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

		By("Deleting PostgresVersion crd")
		err = f.DeletePostgresVersion(postgresVersion.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Test", func() {

		Context("General", func() {

			Context("With PVC", func() {

				It("should run successfully", testGeneralBehaviour)
			})
		})

		Context("Snapshot", func() {

			BeforeEach(func() {
				skipSnapshotDataChecking = false
				snapshot.Spec.DatabaseName = postgres.Name
			})

			Context("In Local", func() {

				BeforeEach(func() {
					skipSnapshotDataChecking = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
				})

				Context("With EmptyDir as Snapshot's backend", func() {
					BeforeEach(func() {
						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						}
					})

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
		})

		Context("Initialize", func() {

			Context("With Script", func() {

				BeforeEach(func() {
					postgres.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/postgres-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", func() {
					// Create Postgres
					createAndWaitForRunning()

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(1))
				})

			})

			Context("With Snapshot", func() {

				var shouldInitializeFromSnapshot = func() {
					// create postgres and take snapshot
					shouldInsertDataAndTakeSnapshot()

					oldPostgres, err := f.GetPostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbagePostgres.Items = append(garbagePostgres.Items, *oldPostgres)

					By("Create postgres from snapshot")
					*postgres = *f.Postgres()
					postgres.Spec.DatabaseSecret = oldPostgres.Spec.DatabaseSecret
					postgres.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					// Create and wait for running Postgres
					createAndWaitForRunning()

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))
				}

				Context("From Local backend", func() {
					var snapPVC *core.PersistentVolumeClaim

					BeforeEach(func() {

						skipSnapshotDataChecking = true
						snapPVC = f.GetPersistentVolumeClaim()
						err := f.CreatePersistentVolumeClaim(snapPVC)
						Expect(err).NotTo(HaveOccurred())

						secret = f.SecretForLocalBackend()
						snapshot.Spec.DatabaseName = postgres.Name
						snapshot.Spec.StorageSecretName = secret.Name

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

					It("should initialize successfully", shouldInitializeFromSnapshot)
				})

				Context("From GCS backend", func() {

					BeforeEach(func() {

						skipSnapshotDataChecking = false
						secret = f.SecretForGCSBackend()
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.DatabaseName = postgres.Name

						snapshot.Spec.GCS = &store.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}
					})

					It("should run successfully", shouldInitializeFromSnapshot)
				})

			})
		})

		Context("Resume", func() {
			var usedInitialized bool

			BeforeEach(func() {
				usedInitialized = false
			})

			var shouldResumeSuccessfully = func() {
				// Create and wait for running Postgres
				createAndWaitForRunning()

				By("Delete postgres")
				f.DeletePostgres(postgres.ObjectMeta)

				By("Wait for postgres to be paused")
				f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

				// Create Postgres object again to resume it
				By("Create Postgres: " + postgres.Name)
				err = f.CreatePostgres(postgres)
				Expect(err).NotTo(HaveOccurred())

				By("Wait for DormantDatabase to be deleted")
				f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

				By("Wait for Running postgres")
				f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

				pg, err := f.GetPostgres(postgres.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				*postgres = *pg
				if usedInitialized {
					_, ok := postgres.Annotations[api.AnnotationInitialized]
					Expect(ok).Should(BeTrue())
				}
			}

			Context("-", func() {
				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("With Init", func() {

				BeforeEach(func() {
					postgres.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/postgres-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("With Snapshot Init", func() {

				BeforeEach(func() {
					skipSnapshotDataChecking = false
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = postgres.Name
				})

				It("should resume successfully", func() {
					// create postgres and take snapshot
					shouldInsertDataAndTakeSnapshot()

					oldPostgres, err := f.GetPostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbagePostgres.Items = append(garbagePostgres.Items, *oldPostgres)

					By("Create postgres from snapshot")
					*postgres = *f.Postgres()
					postgres.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating init Snapshot Postgres without secret name" + postgres.Name)
					err = f.CreatePostgres(postgres)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					postgres.Spec.DatabaseSecret = oldPostgres.Spec.DatabaseSecret
					// Create and wait for running Postgres
					createAndWaitForRunning()

					By("Ping Database")
					f.EventuallyPingDatabase(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

					By("Again delete and resume  " + postgres.Name)

					By("Delete postgres")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for postgres to be paused")
					f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

					// Create Postgres object again to resume it
					By("Create Postgres: " + postgres.Name)
					err = f.CreatePostgres(postgres)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

					By("Wait for Running postgres")
					f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

					postgres, err = f.GetPostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(postgres.Spec.Init).ShouldNot(BeNil())

					By("Ping Database")
					f.EventuallyPingDatabase(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

					By("Checking postgres crd has kubedb.com/initialized annotation")
					_, err = meta_util.GetString(postgres.Annotations, api.AnnotationInitialized)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Resume Multiple times - with init", func() {

				BeforeEach(func() {
					usedInitialized = true
					postgres.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							ScriptPath: "postgres-init-scripts/run.sh",
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/postgres-init-scripts.git",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Postgres
					createAndWaitForRunning()

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")
						By("Delete postgres")
						f.DeletePostgres(postgres.ObjectMeta)

						By("Wait for postgres to be paused")
						f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

						// Create Postgres object again to resume it
						By("Create Postgres: " + postgres.Name)
						err = f.CreatePostgres(postgres)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

						By("Wait for Running postgres")
						f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

						_, err := f.GetPostgres(postgres.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
		})

		Context("SnapshotScheduler", func() {

			BeforeEach(func() {
				secret = f.SecretForLocalBackend()
			})

			Context("With Startup", func() {

				BeforeEach(func() {
					postgres.Spec.BackupSchedule = &api.BackupScheduleSpec{
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

				It("should run scheduler successfully", func() {
					By("Create Secret")
					f.CreateSecret(secret)

					// Create and wait for running Postgres
					createAndWaitForRunning()

					By("Count multiple Snapshot")
					f.EventuallySnapshotCount(postgres.ObjectMeta).Should(matcher.MoreThan(3))
				})
			})

			Context("With Update", func() {
				It("should run scheduler successfully", func() {
					// Create and wait for running Postgres
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update postgres")
					_, err = f.PatchPostgres(postgres.ObjectMeta, func(in *api.Postgres) *api.Postgres {
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
					f.EventuallySnapshotCount(postgres.ObjectMeta).Should(matcher.MoreThan(3))
				})
			})
		})

		Context("Archive with wal-g", func() {

			BeforeEach(func() {
				secret = f.SecretForS3Backend()
				skipWalDataChecking = false
				postgres.Spec.Archiver = &api.PostgresArchiverSpec{
					Storage: &store.Backend{
						StorageSecretName: secret.Name,
						S3: &store.S3Spec{
							Bucket: os.Getenv(S3_BUCKET_NAME),
						},
					},
				}
			})

			Context("Archive and Initialize from wal archive", func() {

				It("should archive and should resume from archive successfully", func() {
					// -- > 1st Postgres < --
					err := f.CreateSecret(secret)
					Expect(err).NotTo(HaveOccurred())

					// Create Postgres
					createAndWaitForRunning()

					By("Creating Schema")
					f.EventuallyCreateSchema(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Creating Table")
					f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

					By("Checking Archive")
					f.EventuallyCountArchive(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					oldPostgres, err := f.GetPostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbagePostgres.Items = append(garbagePostgres.Items, *oldPostgres)

					// -- > 1st Postgres end < --

					// -- > 2nd Postgres < --
					*postgres = *f.Postgres()
					postgres.Spec.Archiver = &api.PostgresArchiverSpec{
						Storage: &store.Backend{
							StorageSecretName: secret.Name,
							S3: &store.S3Spec{
								Bucket: os.Getenv(S3_BUCKET_NAME),
							},
						},
					}

					postgres.Spec.Init = &api.InitSpec{
						PostgresWAL: &api.PostgresWALSourceSpec{
							Backend: store.Backend{
								StorageSecretName: secret.Name,
								S3: &store.S3Spec{
									Bucket: os.Getenv(S3_BUCKET_NAME),
									Prefix: fmt.Sprintf("kubedb/%s/%s/archive/", postgres.Namespace, oldPostgres.Name),
								},
							},
						},
					}

					// Create Postgres
					createAndWaitForRunning()

					By("Ping Database")
					f.EventuallyPingDatabase(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Creating Table")
					f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(6))

					By("Checking Archive")
					f.EventuallyCountArchive(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					oldPostgres, err = f.GetPostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbagePostgres.Items = append(garbagePostgres.Items, *oldPostgres)

					// -- > 2nd Postgres end < --

					// -- > 3rd Postgres < --
					*postgres = *f.Postgres()
					postgres.Spec.Init = &api.InitSpec{
						PostgresWAL: &api.PostgresWALSourceSpec{
							Backend: store.Backend{
								StorageSecretName: secret.Name,
								S3: &store.S3Spec{
									Bucket: os.Getenv(S3_BUCKET_NAME),
									Prefix: fmt.Sprintf("kubedb/%s/%s/archive/", postgres.Namespace, oldPostgres.Name),
								},
							},
						},
					}

					// Create Postgres
					createAndWaitForRunning()

					By("Ping Database")
					f.EventuallyPingDatabase(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(6))
				})
			})

			Context("WipeOut wal data", func() {

				BeforeEach(func() {
					postgres.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should remove wal data from backend", func() {

					err := f.CreateSecret(secret)
					Expect(err).NotTo(HaveOccurred())

					// Create Postgres
					createAndWaitForRunning()

					By("Creating Schema")
					f.EventuallyCreateSchema(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Creating Table")
					f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))

					By("Checking Archive")
					f.EventuallyCountArchive(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

					By("Checking wal data in backend")
					f.EventuallyWalDataFound(postgres).Should(BeTrue())

					By("Deleting Postgres crd")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

					By("Checking Wal data removed from backend")
					f.EventuallyWalDataFound(postgres).Should(BeFalse())
				})
			})
		})

		Context("Termination Policy", func() {

			BeforeEach(func() {
				skipSnapshotDataChecking = false
				secret = f.SecretForGCSBackend()
				snapshot.Spec.StorageSecretName = secret.Name
				snapshot.Spec.GCS = &store.GCSSpec{
					Bucket: os.Getenv(GCS_BUCKET_NAME),
				}
				snapshot.Spec.DatabaseName = postgres.Name
			})

			Context("with TerminationPolicyDoNotTerminate", func() {

				BeforeEach(func() {
					skipSnapshotDataChecking = true
					postgres.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running Postgres
					createAndWaitForRunning()

					By("Delete postgres")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("Postgres is not paused. Check for postgres")
					f.EventuallyPostgres(postgres.ObjectMeta).Should(BeTrue())

					By("Check for Running postgres")
					f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

					By("Update postgres to set spec.terminationPolicy = Pause")
					f.PatchPostgres(postgres.ObjectMeta, func(in *api.Postgres) *api.Postgres {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
				})
			})

			Context("with TerminationPolicyPause (default)", func() {

				It("should create DormantDatabase and resume from it", func() {
					// Run Postgres and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Deleting Postgres crd")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means postgres object is deleted
					By("Waiting for postgres to be paused")
					f.EventuallyDormantDatabaseStatus(postgres.ObjectMeta).Should(matcher.HavePaused())

					By("Checking PVC hasn't been deleted")
					f.EventuallyPVCCount(postgres.ObjectMeta).Should(Equal(1))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(postgres.ObjectMeta).Should(Equal(1))

					By("Checking snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipSnapshotDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					// Create Postgres object again to resume it
					By("Create (resume) Postgres: " + postgres.Name)
					err = f.CreatePostgres(postgres)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

					By("Wait for Running postgres")
					f.EventuallyPostgresRunning(postgres.ObjectMeta).Should(BeTrue())

					By("Checking Table")
					f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))
				})
			})

			Context("with TerminationPolicyDelete", func() {

				BeforeEach(func() {
					postgres.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				AfterEach(func() {
					By("Deleting snapshot: " + snapshot.Name)
					f.DeleteSnapshot(snapshot.ObjectMeta)
				})

				It("should not create DormantDatabase and should not delete secret and snapshot", func() {
					// Run Postgres and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete postgres")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until postgres is deleted")
					f.EventuallyPostgres(postgres.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

					By("Checking PVC has been deleted")
					f.EventuallyPVCCount(postgres.ObjectMeta).Should(Equal(0))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(postgres.ObjectMeta).Should(Equal(1))

					By("Checking Snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipSnapshotDataChecking {
						By("Check for intact snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					postgres.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should not create DormantDatabase and should wipeOut all", func() {
					// Run Postgres and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete postgres")
					err = f.DeletePostgres(postgres.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until postgres is deleted")
					f.EventuallyPostgres(postgres.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(postgres.ObjectMeta).Should(BeFalse())

					By("Checking PVCs has been deleted")
					f.EventuallyPVCCount(postgres.ObjectMeta).Should(Equal(0))

					By("Checking Snapshots has been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())

					By("Checking Secrets has been deleted")
					f.EventuallyDBSecretCount(postgres.ObjectMeta).Should(Equal(0))
				})
			})
		})

		Context("EnvVars", func() {

			Context("With all supported EnvVars", func() {

				It("should create DB with provided EnvVars", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					const (
						dataDir = "/var/pv/pgdata"
						walDir  = "/var/pv/wal"
					)
					dbName = f.App()
					postgres.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  PGDATA,
							Value: dataDir,
						},
						{
							Name:  POSTGRES_DB,
							Value: dbName,
						},
						{
							Name:  POSTGRES_INITDB_ARGS,
							Value: "--data-checksums",
						},
					}

					walEnv := []core.EnvVar{
						{
							Name:  POSTGRES_INITDB_XLOGDIR,
							Value: walDir,
						},
					}
					if framework.DBVersion == "10.2-v1" {
						walEnv = []core.EnvVar{
							{
								Name:  POSTGRES_INITDB_WALDIR,
								Value: walDir,
							},
						}
					}
					postgres.Spec.PodTemplate.Spec.Env = core_util.UpsertEnvVars(postgres.Spec.PodTemplate.Spec.Env, walEnv...)

					// Run Postgres with provided Environment Variables
					testGeneralBehaviour()
				})
			})

			Context("Root Password as EnvVar", func() {

				It("should reject to create Postgres CRD", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					postgres.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  POSTGRES_PASSWORD,
							Value: "not@secret",
						},
					}

					By("Creating Posgres: " + postgres.Name)
					err = f.CreatePostgres(postgres)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("Update EnvVar", func() {

				It("should reject to update EnvVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					postgres.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  POSTGRES_DB,
							Value: dbName,
						},
					}

					// Run Postgres with provided Environment Variables
					testGeneralBehaviour()

					By("Patching EnvVar")
					_, _, err = util.PatchPostgres(f.ExtClient().KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  POSTGRES_DB,
								Value: "patched-db",
							},
						}
						return in
					})
					fmt.Println(err)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("Custom config", func() {

			customConfigs := []string{
				"shared_buffers=256MB",
				"max_connections=300",
			}

			Context("from configMap", func() {
				var userConfig *core.ConfigMap

				BeforeEach(func() {
					userConfig = f.GetCustomConfig(customConfigs)
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					f.DeleteConfigMap(userConfig.ObjectMeta)
				})

				It("should set configuration provided in configMap", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					By("Creating configMap: " + userConfig.Name)
					err := f.CreateConfigMap(userConfig)
					Expect(err).NotTo(HaveOccurred())

					postgres.Spec.ConfigSource = &core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: userConfig.Name,
							},
						},
					}

					// Create Postgres
					createAndWaitForRunning()

					By("Checking postgres configured from provided custom configuration")
					for _, cfg := range customConfigs {
						f.EventuallyPGSettings(postgres.ObjectMeta, dbName, dbUser, cfg).Should(matcher.Use(cfg))
					}
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {
				if skipMessage != "" {
					Skip(skipMessage)
				}
				// Create Postgres
				createAndWaitForRunning()

				By("Creating Schema")
				f.EventuallyCreateSchema(postgres.ObjectMeta, dbName, dbUser).Should(BeTrue())

				By("Creating Table")
				f.EventuallyCreateTable(postgres.ObjectMeta, dbName, dbUser, 3).Should(BeTrue())

				By("Checking Table")
				f.EventuallyCountTable(postgres.ObjectMeta, dbName, dbUser).Should(Equal(3))
			}

			Context("Ephemeral", func() {

				BeforeEach(func() {
					postgres.Spec.StorageType = api.StorageTypeEphemeral
					postgres.Spec.Storage = nil
				})

				Context("General Behaviour", func() {

					BeforeEach(func() {
						postgres.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						postgres.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create Postgres object", func() {
						By("Creating Postgres: " + postgres.Name)
						err := f.CreatePostgres(postgres)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
