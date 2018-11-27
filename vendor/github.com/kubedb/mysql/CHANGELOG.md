# Change Log

## [0.2.0-rc.0](https://github.com/kubedb/mysql/tree/0.2.0-rc.0) (2018-10-15)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.2.0-beta.1...0.2.0-rc.0)

**Merged pull requests:**

- Support custom user & passowrd for backup [\#111](https://github.com/kubedb/mysql/pull/111) ([hossainemruz](https://github.com/hossainemruz))
- Support providing resources for monitoring container [\#110](https://github.com/kubedb/mysql/pull/110) ([tamalsaha](https://github.com/tamalsaha))
- Update kubernetes client libraries to 1.12.0 [\#109](https://github.com/kubedb/mysql/pull/109) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook xray [\#108](https://github.com/kubedb/mysql/pull/108) ([tamalsaha](https://github.com/tamalsaha))
- Various Fixes [\#107](https://github.com/kubedb/mysql/pull/107) ([hossainemruz](https://github.com/hossainemruz))
- Merge ports from service template [\#105](https://github.com/kubedb/mysql/pull/105) ([tamalsaha](https://github.com/tamalsaha))
- Replace doNotPause with TerminationPolicy = DoNotTerminate [\#104](https://github.com/kubedb/mysql/pull/104) ([tamalsaha](https://github.com/tamalsaha))
- Pass resources to NamespaceValidator [\#103](https://github.com/kubedb/mysql/pull/103) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes [\#102](https://github.com/kubedb/mysql/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Support Livecycle hook and container probes [\#101](https://github.com/kubedb/mysql/pull/101) ([tamalsaha](https://github.com/tamalsaha))
- Check if Kubernetes version is supported before running operator [\#100](https://github.com/kubedb/mysql/pull/100) ([tamalsaha](https://github.com/tamalsaha))
- Update package alias [\#99](https://github.com/kubedb/mysql/pull/99) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-beta.1](https://github.com/kubedb/mysql/tree/0.2.0-beta.1) (2018-09-30)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.2.0-beta.0...0.2.0-beta.1)

**Merged pull requests:**

- Revendor api [\#98](https://github.com/kubedb/mysql/pull/98) ([tamalsaha](https://github.com/tamalsaha))
- Fix tests [\#97](https://github.com/kubedb/mysql/pull/97) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api for catalog apigroup [\#96](https://github.com/kubedb/mysql/pull/96) ([tamalsaha](https://github.com/tamalsaha))
- Use --pull flag with docker build \(\#20\) [\#95](https://github.com/kubedb/mysql/pull/95) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0-beta.0](https://github.com/kubedb/mysql/tree/0.2.0-beta.0) (2018-09-20)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0...0.2.0-beta.0)

**Fixed bugs:**

- Search used secrets within same namespace of DB object [\#89](https://github.com/kubedb/mysql/pull/89) ([tamalsaha](https://github.com/tamalsaha))

**Merged pull requests:**

-  Pass extra args to tools.sh [\#93](https://github.com/kubedb/mysql/pull/93) ([the-redback](https://github.com/the-redback))
- Don't try to wipe out Snapshot data for Local backend [\#92](https://github.com/kubedb/mysql/pull/92) ([hossainemruz](https://github.com/hossainemruz))
- Add missing alt-tag docker folder mysql-tools images [\#91](https://github.com/kubedb/mysql/pull/91) ([hossainemruz](https://github.com/hossainemruz))
- Use suffix for updated DBImage & Stop working for deprecated \*Versions [\#90](https://github.com/kubedb/mysql/pull/90) ([hossainemruz](https://github.com/hossainemruz))
- Support Termination Policy [\#88](https://github.com/kubedb/mysql/pull/88) ([hossainemruz](https://github.com/hossainemruz))
- Revendor k8s.io/apiserver [\#87](https://github.com/kubedb/mysql/pull/87) ([tamalsaha](https://github.com/tamalsaha))
- Revendor kubernetes-1.11.3 [\#86](https://github.com/kubedb/mysql/pull/86) ([tamalsaha](https://github.com/tamalsaha))
- Support UpdateStrategy [\#84](https://github.com/kubedb/mysql/pull/84) ([tamalsaha](https://github.com/tamalsaha))
- Add TerminationPolicy for databases [\#83](https://github.com/kubedb/mysql/pull/83) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#82](https://github.com/kubedb/mysql/pull/82) ([tamalsaha](https://github.com/tamalsaha))
- Use IntHash as status.observedGeneration [\#81](https://github.com/kubedb/mysql/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- fix github status [\#80](https://github.com/kubedb/mysql/pull/80) ([tahsinrahman](https://github.com/tahsinrahman))
- update pipeline [\#79](https://github.com/kubedb/mysql/pull/79) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix E2E test for minikube [\#78](https://github.com/kubedb/mysql/pull/78) ([the-redback](https://github.com/the-redback))
- Update pipeline [\#77](https://github.com/kubedb/mysql/pull/77) ([tahsinrahman](https://github.com/tahsinrahman))
- Migrate MySQL [\#75](https://github.com/kubedb/mysql/pull/75) ([tamalsaha](https://github.com/tamalsaha))
- Use official exporter image [\#74](https://github.com/kubedb/mysql/pull/74) ([the-redback](https://github.com/the-redback))
- Update status.ObservedGeneration for failure phase [\#73](https://github.com/kubedb/mysql/pull/73) ([the-redback](https://github.com/the-redback))
- Keep track of ObservedGenerationHash [\#72](https://github.com/kubedb/mysql/pull/72) ([tamalsaha](https://github.com/tamalsaha))
- Use NewObservableHandler [\#71](https://github.com/kubedb/mysql/pull/71) ([tamalsaha](https://github.com/tamalsaha))
- Fix uninstall for concourse [\#70](https://github.com/kubedb/mysql/pull/70) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix uninstall for concourse [\#69](https://github.com/kubedb/mysql/pull/69) ([tahsinrahman](https://github.com/tahsinrahman))
- Revise immutable spec fields [\#68](https://github.com/kubedb/mysql/pull/68) ([tamalsaha](https://github.com/tamalsaha))
- Support passing args via PodTemplate [\#67](https://github.com/kubedb/mysql/pull/67) ([tamalsaha](https://github.com/tamalsaha))
- Introduce storageType : ephemeral [\#66](https://github.com/kubedb/mysql/pull/66) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#65](https://github.com/kubedb/mysql/pull/65) ([tamalsaha](https://github.com/tamalsaha))
- Keep track of observedGeneration in status [\#64](https://github.com/kubedb/mysql/pull/64) ([tamalsaha](https://github.com/tamalsaha))
- Add support for running tests on cncf cluster [\#63](https://github.com/kubedb/mysql/pull/63) ([tahsinrahman](https://github.com/tahsinrahman))
-  Separate StatsService for monitoring [\#62](https://github.com/kubedb/mysql/pull/62) ([the-redback](https://github.com/the-redback))
-  Use MySQLVersion for MySQL images [\#61](https://github.com/kubedb/mysql/pull/61) ([the-redback](https://github.com/the-redback))
- Use updated crd spec [\#60](https://github.com/kubedb/mysql/pull/60) ([tamalsaha](https://github.com/tamalsaha))
- Rename OffshootLabels to OffshootSelectors [\#59](https://github.com/kubedb/mysql/pull/59) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#58](https://github.com/kubedb/mysql/pull/58) ([tamalsaha](https://github.com/tamalsaha))
- Use kmodules monitoring and objectstore api [\#57](https://github.com/kubedb/mysql/pull/57) ([tamalsaha](https://github.com/tamalsaha))
- Refactor concourse scripts [\#56](https://github.com/kubedb/mysql/pull/56) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix command `./hack/make.py test e2e` [\#55](https://github.com/kubedb/mysql/pull/55) ([the-redback](https://github.com/the-redback))
- Set generated binary name to my-operator [\#54](https://github.com/kubedb/mysql/pull/54) ([tamalsaha](https://github.com/tamalsaha))
- Don't add admission/v1beta1 group as a prioritized version [\#53](https://github.com/kubedb/mysql/pull/53) ([tamalsaha](https://github.com/tamalsaha))
- Support custom configuration [\#52](https://github.com/kubedb/mysql/pull/52) ([hossainemruz](https://github.com/hossainemruz))
- Format shell script [\#51](https://github.com/kubedb/mysql/pull/51) ([tamalsaha](https://github.com/tamalsaha))
- Enable status subresource for crds [\#50](https://github.com/kubedb/mysql/pull/50) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v8.0.0 [\#49](https://github.com/kubedb/mysql/pull/49) ([tamalsaha](https://github.com/tamalsaha))
- Fix travis build [\#48](https://github.com/kubedb/mysql/pull/48) ([hossainemruz](https://github.com/hossainemruz))
- Updated osm version to 0.7.1 [\#47](https://github.com/kubedb/mysql/pull/47) ([the-redback](https://github.com/the-redback))
- Support ENV variables in CRDs [\#46](https://github.com/kubedb/mysql/pull/46) ([hossainemruz](https://github.com/hossainemruz))

## [0.1.0](https://github.com/kubedb/mysql/tree/0.1.0) (2018-06-12)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-rc.0...0.1.0)

**Merged pull requests:**

- Fixed missing error return [\#45](https://github.com/kubedb/mysql/pull/45) ([the-redback](https://github.com/the-redback))
- Revendor dependencies [\#44](https://github.com/kubedb/mysql/pull/44) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.0](https://github.com/kubedb/mysql/tree/0.1.0-rc.0) (2018-05-29)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.2...0.1.0-rc.0)

**Merged pull requests:**

- Fix release script [\#43](https://github.com/kubedb/mysql/pull/43) ([tamalsaha](https://github.com/tamalsaha))
- Add changelog [\#42](https://github.com/kubedb/mysql/pull/42) ([tamalsaha](https://github.com/tamalsaha))
- Concourse [\#41](https://github.com/kubedb/mysql/pull/41) ([tahsinrahman](https://github.com/tahsinrahman))
- Fixed kubeconfig plugin for Cloud Providers && Storage is required for MySQL [\#40](https://github.com/kubedb/mysql/pull/40) ([the-redback](https://github.com/the-redback))
- Remove lost+found directory before initializing mysql [\#39](https://github.com/kubedb/mysql/pull/39) ([the-redback](https://github.com/the-redback))
- Refactored E2E testing to support E2E testing with admission webhook in cloud [\#38](https://github.com/kubedb/mysql/pull/38) ([the-redback](https://github.com/the-redback))
- Skip delete requests for empty resources [\#37](https://github.com/kubedb/mysql/pull/37) ([the-redback](https://github.com/the-redback))
- Don't panic if admission options is nil [\#36](https://github.com/kubedb/mysql/pull/36) ([tamalsaha](https://github.com/tamalsaha))
- Disable admission controllers for webhook server [\#35](https://github.com/kubedb/mysql/pull/35) ([tamalsaha](https://github.com/tamalsaha))
- Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 [\#34](https://github.com/kubedb/mysql/pull/34) ([the-redback](https://github.com/the-redback))
- Update client-go to 7.0.0 [\#33](https://github.com/kubedb/mysql/pull/33) ([tamalsaha](https://github.com/tamalsaha))
- Added update script of docker for mysql-tools:8 [\#32](https://github.com/kubedb/mysql/pull/32) ([the-redback](https://github.com/the-redback))
- Added support of mysql:5.7 [\#31](https://github.com/kubedb/mysql/pull/31) ([the-redback](https://github.com/the-redback))
- Add support for one informer and N-eventHandler for snapshot, dromantdb and Job [\#30](https://github.com/kubedb/mysql/pull/30) ([the-redback](https://github.com/the-redback))
- Use metrics from kube apiserver [\#29](https://github.com/kubedb/mysql/pull/29) ([tamalsaha](https://github.com/tamalsaha))
- Bundle webhook server and Use SharedInformerFactory [\#28](https://github.com/kubedb/mysql/pull/28) ([the-redback](https://github.com/the-redback))
- Move MySQL AdmissionWebhook packages into MySQL repository [\#27](https://github.com/kubedb/mysql/pull/27) ([the-redback](https://github.com/the-redback))
- Use mysql:8.0.3 image as mysql:8.0 [\#26](https://github.com/kubedb/mysql/pull/26) ([the-redback](https://github.com/the-redback))
- Add travis yaml [\#25](https://github.com/kubedb/mysql/pull/25) ([tahsinrahman](https://github.com/tahsinrahman))

## [0.1.0-beta.2](https://github.com/kubedb/mysql/tree/0.1.0-beta.2) (2018-02-27)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.1...0.1.0-beta.2)

**Merged pull requests:**

- Migrating to apps/v1 [\#23](https://github.com/kubedb/mysql/pull/23) ([the-redback](https://github.com/the-redback))
- update validation [\#22](https://github.com/kubedb/mysql/pull/22) ([aerokite](https://github.com/aerokite))
-  Fix dormantDB matching: pass same type to Equal method [\#21](https://github.com/kubedb/mysql/pull/21) ([the-redback](https://github.com/the-redback))
- Use official code generator scripts [\#20](https://github.com/kubedb/mysql/pull/20) ([tamalsaha](https://github.com/tamalsaha))
-  Fixed dormantdb matching & Raised throttling time & Fixed MySQL version Checking [\#19](https://github.com/kubedb/mysql/pull/19) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.1](https://github.com/kubedb/mysql/tree/0.1.0-beta.1) (2018-01-29)
[Full Changelog](https://github.com/kubedb/mysql/compare/0.1.0-beta.0...0.1.0-beta.1)

**Merged pull requests:**

- converted to k8s 1.9 & Improved InitSpec in DormantDB & Added support for Job watcher & Improved Tests [\#17](https://github.com/kubedb/mysql/pull/17) ([the-redback](https://github.com/the-redback))
- Fixed logger, analytics and Removed rbac stuff [\#16](https://github.com/kubedb/mysql/pull/16) ([the-redback](https://github.com/the-redback))
- Add rbac stuffs for mysql-exporter [\#15](https://github.com/kubedb/mysql/pull/15) ([the-redback](https://github.com/the-redback))
-  Review Mysql docker images and Fixed monitring [\#14](https://github.com/kubedb/mysql/pull/14) ([the-redback](https://github.com/the-redback))

## [0.1.0-beta.0](https://github.com/kubedb/mysql/tree/0.1.0-beta.0) (2018-01-07)
**Merged pull requests:**

- Rename ms-operator to my-operator [\#13](https://github.com/kubedb/mysql/pull/13) ([tamalsaha](https://github.com/tamalsaha))
- Fix Analytics and pass client-id as ENV to Snapshot Job [\#12](https://github.com/kubedb/mysql/pull/12) ([the-redback](https://github.com/the-redback))
- Add docker-registry and WorkQueue  [\#10](https://github.com/kubedb/mysql/pull/10) ([the-redback](https://github.com/the-redback))
- Set client id for analytics [\#9](https://github.com/kubedb/mysql/pull/9) ([tamalsaha](https://github.com/tamalsaha))
- Fix CRD Registration [\#8](https://github.com/kubedb/mysql/pull/8) ([the-redback](https://github.com/the-redback))
- Update pkg paths to kubedb org [\#7](https://github.com/kubedb/mysql/pull/7) ([tamalsaha](https://github.com/tamalsaha))
- Assign default Prometheus Monitoring Port [\#6](https://github.com/kubedb/mysql/pull/6) ([the-redback](https://github.com/the-redback))
- mysql-util docker image [\#5](https://github.com/kubedb/mysql/pull/5) ([the-redback](https://github.com/the-redback))
- Add Snapshot Backup, Restore and Backup-Scheduler [\#4](https://github.com/kubedb/mysql/pull/4) ([the-redback](https://github.com/the-redback))
- Update ./hack folder [\#3](https://github.com/kubedb/mysql/pull/3) ([tamalsaha](https://github.com/tamalsaha))
- Mysql db - Inititalizing  [\#2](https://github.com/kubedb/mysql/pull/2) ([the-redback](https://github.com/the-redback))
- Add skeleton for mysql [\#1](https://github.com/kubedb/mysql/pull/1) ([aerokite](https://github.com/aerokite))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*