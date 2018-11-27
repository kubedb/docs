package leader_election

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func pgConnString(hostname string) string {

	//hostname := getEnv("PRIMARY_HOST", "localhost")
	username := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")

	info := fmt.Sprintf("host=%s port=%d dbname=%s "+
		"sslmode=%s user=%s password=%s ",
		hostname,
		5432,
		"postgres",
		"disable",
		username,
		password,
	)

	return info
}

func setPosgresUserPassword(username, password string) {
	log.Printf("Trying to set password to Postgres user: %s", username)

	if db, err := sql.Open("postgres", pgConnString("localhost")); db != nil {
		defer db.Close()

		if _, err = db.Exec("ALTER USER $1 WITH PASSWORD $2", username, password); err == nil {
			log.Printf("Password successfully set to %s", password)
		}
		log.Println("query error")
	} else {
		log.Println("connection error")
	}
}

func isPostgresOnline(ctx context.Context, hostname string, wait bool) bool {
	// authung! dangerous function
	//if wait == true function will wait until connection established
	returnValue := false

	for {
		select {
		case <-ctx.Done():
			return returnValue
		default:
		}

		log.Println("Checking connection to master")

		if db, err := sql.Open("postgres", pgConnString(hostname)); db != nil {
			defer db.Close()
			if _, err = db.Exec("SELECT 1;"); err == nil {
				returnValue = true
				break
			}
			log.Println("query error")
			db.Close()
		} else {
			log.Println("connection error")
		}
		if !wait {
			break
		}
		time.Sleep(time.Second * 60)
	}
	return returnValue
}

func dataDirectoryCleanup() {
	log.Println("dataDirectoryCleanup: Removing and creating data directory")
	PGDATA := getEnv("PGDATA", "/var/pv/data")
	os.RemoveAll(PGDATA)
	os.MkdirAll(PGDATA, 0755)
	setPermission()
}

func dataDirectoryCreateAfterWalg() {
	log.Println("dataDirectoryCreateAfterWalg: creating directories after wal-g")
	PGDATA := getEnv("PGDATA", "/var/pv/data")
	dirList := []string{
		"/pg_tblspc",
		"/pg_twophase",
		"/pg_stat",
		"/pg_commit_ts",
		"/pg_logical/snapshots",
		"/pg_logical/mappings",
	}
	for _, dir := range dirList {
		os.MkdirAll(PGDATA+dir, 0700)
	}
}

func execWalgAction(ctx context.Context, walgCommand string, params ...string) error {
	var env []string
	env = append(env, fmt.Sprintf("WALE_S3_PREFIX=%s", getEnv("ARCHIVE_S3_PREFIX", "")))
	// auth for wal-g
	env = append(env, fmt.Sprintf("PGUSER=%s", getEnv("POSTGRES_USER", "")))
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", getEnv("POSTGRES_PASSWORD", "")))

	awsKeyFile := "/srv/wal-g/archive/secrets/AWS_ACCESS_KEY_ID"
	awsSecretFile := "/srv/wal-g/archive/secrets/AWS_SECRET_ACCESS_KEY"

	awsKey, err := ioutil.ReadFile(awsKeyFile)
	// aws key file ansent
	if err != nil {
		log.Println(err)
	}

	awsSecret, err := ioutil.ReadFile(awsSecretFile)
	// aws secret file absent
	if err != nil {
		log.Println(err)
	}
	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", awsKey))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", awsSecret))
	// need to forward "wal-g", walgCommand, params...
	arg := []string{"root", "wal-g"}
	arg = append(arg, walgCommand)
	arg = append(arg, params...)
	return runCmd(ctx, env, "su-exec", arg...)
}

func execBaseBackup(ctx context.Context) error {
	log.Println("execBaseBackup: running pg_basebackup")
	var env []string
	env = append(env, fmt.Sprintf("PGUSER=%s", getEnv("POSTGRES_USER", "postgres")))
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", getEnv("POSTGRES_PASSWORD", "postgres")))
	pgdata := getEnv("PGDaTA", "/var/pv/data")
	pghost := fmt.Sprintf("--host=%s", getEnv("PRIMARY_HOST", ""))
	return runCmd(ctx, env, "pg_basebackup", "-X", "fetch", "--no-password", "--pgdata", pgdata, pghost)
}

func execPostgresAction(ctx context.Context, action string) {
	log.Printf("execPostgresAction: %s", action)
	var env []string
	env = append(env, fmt.Sprintf("WALE_S3_PREFIX=%s", getEnv("ARCHIVE_S3_PREFIX", "")))
	// auth for wal-g
	env = append(env, fmt.Sprintf("PGUSER=%s", getEnv("POSTGRES_USER", "")))
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", getEnv("POSTGRES_PASSWORD", "")))

	awsKeyFile := "/srv/wal-g/archive/secrets/AWS_ACCESS_KEY_ID"
	awsSecretFile := "/srv/wal-g/archive/secrets/AWS_SECRET_ACCESS_KEY"

	awsKey, err := ioutil.ReadFile(awsKeyFile)
	// aws key file ansent
	if err != nil {
		log.Printf("Error opening file %v", err)
	}

	awsSecret, err := ioutil.ReadFile(awsSecretFile)
	// aws secret file absent
	if err != nil {
		log.Printf("Error opening file %v", err)
	}
	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", awsKey))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", awsSecret))
	err = runCmd(ctx, env, "su-exec", "postgres", "pg_ctl", "-D", getEnv("PGDATA", "/var/pv/data"), "-w", action)
	if err != nil {
		log.Printf("Error happened during pg_ctl: %v", err)
	}

}

func postgresMakeEmptyDB(ctx context.Context) {
	log.Println("postgresMakeEmptyDB: Create empty database for postgres")
	var env []string
	err := runCmd(ctx, env, "initdb", fmt.Sprintf("--pgdata=%s", getEnv("PGDATA", "/var/pv/data")))
	if err != nil {
		log.Printf("Error happened during initdb: %v", err)
	}

}

func postgresMakeConfigs(role string) {
	log.Printf("Create config files for postgres, role: %s", role)
	if role == RolePrimary {
		var env []string
		// copy template to /tmp
		err := runCmd(context.TODO(), env, "cp", "/scripts/primary/postgresql.conf", "/tmp/")
		if err != nil {
			log.Printf("Error happened during cp /scripts/primary/postgresql.conf: %v", err)
		}

		// append config
		lines := []string{
			"wal_level = replica",
			"max_wal_senders = 99",
			"wal_keep_segments = 32",
		}
		if getEnv("STREAMING", "") == "synchronous" {
			// setup synchronous streaming replication
			lines = append(lines, "synchronous_commit = remote_write")
			lines = append(lines, "synchronous_standby_names = '*'")
		}
		if getEnv("ARCHIVE", "") == "wal-g" {
			lines = append(lines, "archive_command = 'wal-g wal-push %p'")
			lines = append(lines, "archive_timeout = 60")
			lines = append(lines, "archive_mode = always")
		}
		err = appendFile("/tmp/postgresql.conf", lines)
		if err != nil {
			log.Printf("Error happened during appendFile /tmp/postgresql.conf: %v", err)
		}

		// move configs to PGDATA
		err = runCmd(context.TODO(), env, "mv", "/tmp/postgresql.conf", getEnv("PGDATA", "/var/pv/data"))
		if err != nil {
			log.Printf("Error happened during mv /tmp/postgresql.conf: %v", err)
		}
		err = runCmd(context.TODO(), env, "mv", "/scripts/primary/pg_hba.conf", getEnv("PGDATA", "/var/pv/data"))
		if err != nil {
			log.Printf("Error happened during mv /scripts/primary/pg_hba.conf: %v", err)
		}
	}
	if role == RoleReplica {
		var env []string
		// copy template to /tmp
		err := runCmd(context.TODO(), env, "cp", "/scripts/replica/recovery.conf", "/tmp/")
		if err != nil {
			log.Printf("Error happened during cp /scripts/replica/recovery.conf: %v", err)
		}

		// append recovery.conf
		lines := []string{
			"recovery_target_timeline = 'latest'",
			fmt.Sprintf("archive_cleanup_command = 'pg_archivecleanup %s %r'", getEnv("PGWAL", "")),
			fmt.Sprintf("primary_conninfo = 'application_name=%s host=%s'", getEnv("HOSTNAME", ""), getEnv("PRIMARY_HOST", "")),
		}
		err = appendFile("/tmp/recovery.conf", lines)
		if err != nil {
			log.Printf("Error happened during appendFile /tmp/recovery.conf: %v", err)
		}

		// append postgresql.conf
		err = runCmd(context.TODO(), env, "cp", "/scripts/primary/postgresql.conf", "/tmp/")
		if err != nil {
			log.Printf("Error happened during cp /scripts/primary/postgresql.conf: %v", err)
		}
		lines = []string{
			"wal_level = replica",
			"max_wal_senders = 99",
			"wal_keep_segments = 32",
		}
		if getEnv("STANDBY", "") == "hot" {
			lines = append(lines, "hot_standby = on")
		}
		if getEnv("STREAMING", "") == "synchronous" {
			// setup synchronous streaming replication
			lines = append(lines, "synchronous_commit = remote_write")
			lines = append(lines, "synchronous_standby_names = '*'")
		}
		if getEnv("ARCHIVE", "") == "wal-g" {
			lines = append(lines, "archive_command = 'wal-g wal-push %p'")
			lines = append(lines, "archive_timeout = 60")
			lines = append(lines, "archive_mode = always")
		}
		err = appendFile("/tmp/postgresql.conf", lines)
		if err != nil {
			log.Printf("Error happened during appendFile /tmp/postgresql.conf: %v", err)
		}

		// move configs to PGDATA
		err = runCmd(context.TODO(), env, "mv", "/tmp/postgresql.conf", getEnv("PGDATA", "/var/pv/data"))
		if err != nil {
			log.Printf("Error happened during mv /tmp/postgresql.conf: %v", err)
		}

		err = runCmd(context.TODO(), env, "mv", "/tmp/recovery.conf", getEnv("PGDATA", "/var/pv/data"))
		if err != nil {
			log.Printf("Error happened during mv /tmp/recovery.conf: %v", err)
		}

		err = runCmd(context.TODO(), env, "mv", "/scripts/primary/pg_hba.conf", getEnv("PGDATA", "/var/pv/data"))
		if err != nil {
			log.Printf("Error happened during mv /scripts/primary/pg_hba.conf: %v", err)
		}

	}
}

func restoreMasterFromBackup(ctx context.Context) error {
	// absolutely clean data directory
	dataDirectoryCleanup()
	// some actions to start recovery
	err := execWalgAction(context.TODO(), "backup-list")
	if err != nil {
		log.Printf("Error happened during execWalgAction: %v", err)
	}

	restoreComplete := execWalgAction(ctx, "backup-fetch", getEnv("PGDATA", "/var/pv/data"), getEnv("BACKUP_NAME", "LATEST"))

	dataDirectoryCreateAfterWalg()
	postgresMakeConfigs(RolePrimary)
	// additional lines to recovery.conf
	lines := []string{}
	if getEnv("PITR", "") != "" {

		target_inclusive := getEnv("TARGET_INCLUSIVE", "true")
		target_time := getEnv("TARGET_TIME", "")
		target_timeline := getEnv("TARGET_TIMELINE", "")
		target_xid := getEnv("TARGET_XID", "")

		lines = []string{
			fmt.Sprintf("recovery_target_inclusive = '%s'", target_inclusive),
			"recovery_target_action = 'promote'",
		}
		if target_time != "" {
			lines = append(lines, fmt.Sprintf("recovery_target_time = '%s'", target_time))
		}
		if target_timeline != "" {
			lines = append(lines, fmt.Sprintf("recovery_target_timeline = '%s'", target_timeline))
		}
		if target_xid != "" {
			lines = append(lines, fmt.Sprintf("recovery_target_xid = '%s'", target_xid))
		}

	}
	lines = append(lines, "restore_command = 'wal-g wal-fetch %f %p'")
	err = appendFile(getEnv("PGDATA", "/var/pv/data")+"/recovery.conf", lines)
	if err != nil {
		log.Printf("Error happened during appendFile recovery.conf: %v", err)
	}

	os.Remove(getEnv("PGDATA", "/var/pv/data") + "/recovery.done")
	setPermission()
	postgresContext, _ := context.WithCancel(ctx)
	go execPostgresAction(postgresContext, "start")
	// backup done, start Postgres
	return restoreComplete
}

func waitForRecoveryDone(ctx context.Context) {
	recovery_done_file := getEnv("PGDATA", "/var/pv/data") + "/recovery.done"
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if _, err := os.Stat(recovery_done_file); !os.IsNotExist(err) {
				return
			}
			log.Println("master loop: Waiting recovery.done to be created")
			time.Sleep(time.Second)
		}
	}
}
