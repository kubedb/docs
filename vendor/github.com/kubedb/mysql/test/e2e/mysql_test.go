package e2e_test

import (
	"fmt"
	"os"

	"github.com/appscode/go/log"
	meta_util "github.com/appscode/kutil/meta"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/mysql/test/e2e/framework"
	"github.com/kubedb/mysql/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	store "kmodules.xyz/objectstore-api/api/v1"
)

const (
	S3_BUCKET_NAME       = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME      = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME = "SWIFT_CONTAINER_NAME"
	MYSQL_DATABASE       = "MYSQL_DATABASE"
	MYSQL_ROOT_PASSWORD  = "MYSQL_ROOT_PASSWORD"
)

var _ = Describe("MySQL", func() {
	var (
		err              error
		f                *framework.Invocation
		mysql            *api.MySQL
		garbageMySQL     *api.MySQLList
		mysqlVersion     *catalog.MySQLVersion
		snapshot         *api.Snapshot
		secret           *core.Secret
		skipMessage      string
		skipDataChecking bool
		dbName           string
	)

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQL()
		garbageMySQL = new(api.MySQLList)
		mysqlVersion = f.MySQLVersion()
		snapshot = f.Snapshot()
		skipMessage = ""
		skipDataChecking = true
		dbName = "mysql"
	})

	var createAndWaitForRunning = func() {
		By("Create MySQLVersion: " + mysqlVersion.Name)
		err = f.CreateMySQLVersion(mysqlVersion)
		Expect(err).NotTo(HaveOccurred())

		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(mysql.ObjectMeta, dbName).Should(BeTrue())
	}

	var testGeneralBehaviour = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}
		// Create MySQL
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

		By("Inserting Rows")
		f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for mysql to be paused")
		f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

		// Create MySQL object again to resume it
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for DormantDatabase to be deleted")
		f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
	}

	var shouldTakeSnapshot = func() {
		// Create and wait for running MySQL
		createAndWaitForRunning()

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeed snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var shouldInsertDataAndTakeSnapshot = func() {
		// Create and wait for running MySQL
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

		By("Inserting Row")
		f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeed snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var deleteTestResource = func() {
		if mysql == nil {
			log.Infoln("Skipping cleanup. Reason: mysql is nil")
			return
		}

		By("Check if mysql " + mysql.Name + " exists.")
		my, err := f.GetMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// MySQL was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: MySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if my.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for mysql to be paused")
			f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut mysql")
			_, err := f.PatchDormantDatabase(mysql.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(mysql.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for mysql resources to be wipedOut")
		f.EventuallyWipedOut(mysql.ObjectMeta).Should(Succeed())
	}

	var deleteSnapshot = func() {

		By("Deleting Snapshot: " + snapshot.Name)
		err = f.DeleteSnapshot(snapshot.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		if !skipDataChecking {
			// do not try to check snapshot data if secret does not exist
			_, err = f.GetSecret(secret.ObjectMeta)
			if err != nil && kerr.IsNotFound(err) {
				log.Infof("Skipping checking snapshot data. Reason: secret %s not found", secret.Name)
				return
			}
			Expect(err).NotTo(HaveOccurred())

			By("Checking Snapshot's data wiped out from backend")
			f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
		}
	}

	AfterEach(func() {
		// delete resources for current MySQL
		deleteTestResource()

		// old MySQL are in garbageMySQL list. delete their resources.
		for _, my := range garbageMySQL.Items {
			*mysql = my
			deleteTestResource()
		}

		By("Deleting MySQLVersion crd")
		err := f.DeleteMySQLVersion(mysqlVersion.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers()
	})

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", testGeneralBehaviour)
			})
		})

		Context("Snapshot", func() {

			BeforeEach(func() {
				skipDataChecking = false
				snapshot.Spec.DatabaseName = mysql.Name
			})

			AfterEach(func() {
				// delete snapshot and check for data wipeOut
				deleteSnapshot()

				By("Deleting secret: " + secret.Name)
				err := f.DeleteSecret(secret.ObjectMeta)
				if err != nil && !kerr.IsNotFound(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			Context("In Local", func() {

				BeforeEach(func() {
					skipDataChecking = true
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
						err := f.DeleteSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for Snapshot to be deleted")
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

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
				})

				Context("Without Init", func() {
					It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
				})

				Context("With Init", func() {
					BeforeEach(func() {
						mysql.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/mysql-init-scripts.git",
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
						mysql.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/mysql-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create MySQL and take Snapshot
						shouldTakeSnapshot()

						oldSnapshot := snapshot

						// create new Snapshot
						snapshot := f.Snapshot()
						snapshot.Spec.DatabaseName = mysql.Name
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.GCS = &store.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataChecking {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						By(fmt.Sprintf("Delete Snapshot %v", snapshot.Name))
						err = f.DeleteSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for Snapshot to be deleted")
						f.EventuallySnapshot(mysql.ObjectMeta).Should(BeFalse())
						if !skipDataChecking {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
						}

						snapshot = oldSnapshot

						By(fmt.Sprintf("Checking old Snapshot %v still exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataChecking {
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
						Container: os.Getenv(AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &store.SwiftSpec{
						Container: os.Getenv(SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
			})
		})

		Context("Initialize", func() {

			Context("With Script", func() {
				BeforeEach(func() {
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", func() {
					// Create MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
				})
			})

			Context("With Snapshot", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				var shouldInitializeFromSnapshot = func() {
					// Create MySQL and take Snapshot
					shouldInsertDataAndTakeSnapshot()

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)

					By("Create mysql from snapshot")
					mysql = f.MySQL()
					mysql.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating init Snapshot Mysql without secret name" + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret

					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
				}

				Context("From Local backend", func() {
					var snapPVC *core.PersistentVolumeClaim

					BeforeEach(func() {

						skipDataChecking = true
						snapPVC = f.GetPersistentVolumeClaim()
						err := f.CreatePersistentVolumeClaim(snapPVC)
						Expect(err).NotTo(HaveOccurred())

						secret = f.SecretForLocalBackend()
						snapshot.Spec.DatabaseName = mysql.Name
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

						skipDataChecking = false
						secret = f.SecretForGCSBackend()
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.DatabaseName = mysql.Name

						snapshot.Spec.GCS = &store.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}
					})

					It("should initialize successfully", shouldInitializeFromSnapshot)
				})
			})
		})

		Context("Resume", func() {

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					// Delete without caring if DB is resumed
					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for MySQL to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
				})
			})

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
				})
			})

			Context("with init Script", func() {
				BeforeEach(func() {
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					mysql, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(mysql.Spec.Init).NotTo(BeNil())

					By("Checking MySQL crd does not have kubedb.com/initialized annotation")
					_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("With Snapshot Init", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				BeforeEach(func() {
					skipDataChecking = false
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = mysql.Name
				})

				It("should resume successfully", func() {
					// Create MySQL and take Snapshot
					shouldInsertDataAndTakeSnapshot()

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)

					By("Create mysql from snapshot")
					mysql = f.MySQL()
					mysql.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating MySQL without secret name to init from Snapshot: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret

					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					mysql, err = f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(mysql.Spec.Init).ShouldNot(BeNil())

					By("Checking MySQL has kubedb.com/initialized annotation")
					_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Multiple times with init", func() {

				BeforeEach(func() {
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")

						By("Delete mysql")
						err = f.DeleteMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for mysql to be paused")
						f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

						// Create MySQL object again to resume it
						By("Create MySQL: " + mysql.Name)
						err = f.CreateMySQL(mysql)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

						By("Wait for Running mysql")
						f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

						By("Checking Row Count of Table")
						f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

						mysql, err := f.GetMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
						Expect(mysql.Spec.Init).ShouldNot(BeNil())

						By("Checking MySQL crd does not have kubedb.com/initialized annotation")
						_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
						Expect(err).To(HaveOccurred())
					}
				})
			})
		})

		Context("SnapshotScheduler", func() {

			BeforeEach(func() {
				skipDataChecking = false
			})

			AfterEach(func() {
				snapshotList, err := f.GetSnapshotList(mysql.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				for _, snap := range snapshotList.Items {
					snapshot = &snap

					// delete snapshot and check for data wipeOut
					deleteSnapshot()
				}

				By("Deleting secret: " + secret.Name)
				err = f.DeleteSecret(secret.ObjectMeta)
				if err != nil && !kerr.IsNotFound(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			Context("With Startup", func() {

				var shouldStartupSchedular = func() {
					By("Create Secret")
					f.CreateSecret(secret)

					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
				}

				Context("with local", func() {
					BeforeEach(func() {
						skipDataChecking = true
						secret = f.SecretForLocalBackend()
						mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
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

					It("should run scheduler successfully", shouldStartupSchedular)
				})

				Context("with GCS", func() {
					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 1m",
							Backend: store.Backend{
								StorageSecretName: secret.Name,
								GCS: &store.GCSSpec{
									Bucket: os.Getenv(GCS_BUCKET_NAME),
								},
							},
						}
					})

					It("should run scheduler successfully", shouldStartupSchedular)
				})
			})

			Context("With Update - with Local", func() {

				BeforeEach(func() {
					skipDataChecking = true
					secret = f.SecretForLocalBackend()
				})

				It("should run scheduler successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update mysql")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
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

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
				})
			})

			Context("Re-Use DormantDatabase's scheduler", func() {

				BeforeEach(func() {
					skipDataChecking = true
					secret = f.SecretForLocalBackend()
				})

				It("should re-use scheduler successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update mysql")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
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

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(5))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
				})
			})
		})

		Context("Termination Policy", func() {

			BeforeEach(func() {
				skipDataChecking = false
				secret = f.SecretForGCSBackend()
				snapshot.Spec.StorageSecretName = secret.Name
				snapshot.Spec.GCS = &store.GCSSpec{
					Bucket: os.Getenv(GCS_BUCKET_NAME),
				}
				snapshot.Spec.DatabaseName = mysql.Name
			})

			Context("with TerminationDoNotTerminate", func() {
				BeforeEach(func() {
					skipDataChecking = true
					mysql.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("MySQL is not paused. Check for mysql")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeTrue())

					By("Check for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Update mysql to set spec.terminationPolicy = Pause")
					f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
				})
			})

			Context("with TerminationPolicyPause (default)", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should create DormantDatabase and resume from it", func() {
					// Run MySQL and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Deleting MySQL crd")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means mysql object is deleted
					By("Waiting for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					By("Checking PVC hasn't been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(1))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))

					By("Checking snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					// Create MySQL object again to resume it
					By("Create (resume) MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking row count of table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
				})
			})

			Context("with TerminationPolicyDelete", func() {

				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should not create DormantDatabase and should not delete secret and snapshot", func() {
					// Run MySQL and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until mysql is deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVC has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))

					By("Checking Snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipDataChecking {
						By("Check for intact snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should not create DormantDatabase and should wipeOut all", func() {
					// Run MySQL and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until mysql is deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVCs has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

					By("Checking Snapshots has been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())

					By("Checking Secrets has been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(0))
				})
			})
		})

		Context("EnvVars", func() {

			Context("Database Name as EnvVar", func() {

				It("should create DB with name provided in EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()
				})
			})

			Context("Root Password as EnvVar", func() {

				It("should reject to create MySQL CRD", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_ROOT_PASSWORD,
							Value: "not@secret",
						},
					}
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("Update EnvVar", func() {

				It("should reject to update EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()

					By("Patching EnvVar")
					_, _, err = util.PatchMySQL(f.ExtClient().KubedbV1alpha1(), mysql, func(in *api.MySQL) *api.MySQL {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  MYSQL_DATABASE,
								Value: "patched-db",
							},
						}
						return in
					})
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("Custom config", func() {

			customConfigs := []string{
				"max_connections=200",
				"read_buffer_size=1048576", // 1MB
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

					mysql.Spec.ConfigSource = &core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: userConfig.Name,
							},
						},
					}

					// Create MySQL
					createAndWaitForRunning()

					By("Checking mysql configured from provided custom configuration")
					for _, cfg := range customConfigs {
						f.EventuallyMySQLVariable(mysql.ObjectMeta, dbName, cfg).Should(matcher.UseCustomConfig(cfg))
					}
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create MySQL
				createAndWaitForRunning()

				By("Creating Table")
				f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

				By("Inserting Rows")
				f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

				By("Checking Row Count of Table")
				f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))
			}

			Context("Ephemeral", func() {

				Context("General Behaviour", func() {

					BeforeEach(func() {
						mysql.Spec.StorageType = api.StorageTypeEphemeral
						mysql.Spec.Storage = nil
						mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						mysql.Spec.StorageType = api.StorageTypeEphemeral
						mysql.Spec.Storage = nil
						mysql.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create MySQL object", func() {

						By("Creating MySQL: " + mysql.Name)
						err := f.CreateMySQL(mysql)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
