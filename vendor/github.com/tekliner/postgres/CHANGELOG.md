# Change Log

## [0.9.0-rc.0](https://github.com/kubedb/postgres/tree/0.9.0-rc.0) (2018-10-15)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.9.0-beta.1...0.9.0-rc.0)

**Merged pull requests:**

- Fix build [\#222](https://github.com/kubedb/postgres/pull/222) ([tamalsaha](https://github.com/tamalsaha))
- Fix build [\#221](https://github.com/kubedb/postgres/pull/221) ([tamalsaha](https://github.com/tamalsaha))
- Support providing resources for monitoring container [\#220](https://github.com/kubedb/postgres/pull/220) ([hossainemruz](https://github.com/hossainemruz))
- Recognize denied request by any webhook in xray [\#219](https://github.com/kubedb/postgres/pull/219) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes [\#217](https://github.com/kubedb/postgres/pull/217) ([hossainemruz](https://github.com/hossainemruz))
- Update kubernetes client libraries to 1.12.0 [\#216](https://github.com/kubedb/postgres/pull/216) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook xray [\#215](https://github.com/kubedb/postgres/pull/215) ([tamalsaha](https://github.com/tamalsaha))
- Merge ports from service template [\#213](https://github.com/kubedb/postgres/pull/213) ([tamalsaha](https://github.com/tamalsaha))
- Remove remaining DoNotPause [\#212](https://github.com/kubedb/postgres/pull/212) ([tamalsaha](https://github.com/tamalsaha))
- Set TerminationPolicy to WipeOut in e2e tests [\#211](https://github.com/kubedb/postgres/pull/211) ([tamalsaha](https://github.com/tamalsaha))
- Replace doNotPause with TerminationPolicy = DoNotTerminate [\#210](https://github.com/kubedb/postgres/pull/210) ([tamalsaha](https://github.com/tamalsaha))
- Pass resources to NamespaceValidator [\#209](https://github.com/kubedb/postgres/pull/209) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook for Namespace deletion [\#208](https://github.com/kubedb/postgres/pull/208) ([tamalsaha](https://github.com/tamalsaha))
- Use FQDN for kube-apiserver in AKS [\#207](https://github.com/kubedb/postgres/pull/207) ([tamalsaha](https://github.com/tamalsaha))
- Support Livecycle hook and container probes [\#206](https://github.com/kubedb/postgres/pull/206) ([tamalsaha](https://github.com/tamalsaha))
- Check if Kubernetes version is supported before running operator [\#205](https://github.com/kubedb/postgres/pull/205) ([tamalsaha](https://github.com/tamalsaha))

## [0.9.0-beta.1](https://github.com/kubedb/postgres/tree/0.9.0-beta.1) (2018-09-30)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.9.0-beta.0...0.9.0-beta.1)

**Merged pull requests:**

- Revendor api [\#204](https://github.com/kubedb/postgres/pull/204) ([tamalsaha](https://github.com/tamalsaha))
- Change streaming mode constants to CamelCase [\#203](https://github.com/kubedb/postgres/pull/203) ([tamalsaha](https://github.com/tamalsaha))
- Fix tests [\#202](https://github.com/kubedb/postgres/pull/202) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api for catalog apigroup [\#201](https://github.com/kubedb/postgres/pull/201) ([tamalsaha](https://github.com/tamalsaha))
- Cherry Pick into 'release-0.9': Fix missing '--' argument receiver on postgres-tools \(\#199\) [\#200](https://github.com/kubedb/postgres/pull/200) ([the-redback](https://github.com/the-redback))
- Fix missing '--' argument receiver on postgres-tools [\#199](https://github.com/kubedb/postgres/pull/199) ([the-redback](https://github.com/the-redback))
- Use --pull flag with docker build \(\#20\) [\#198](https://github.com/kubedb/postgres/pull/198) ([tamalsaha](https://github.com/tamalsaha))

## [0.9.0-beta.0](https://github.com/kubedb/postgres/tree/0.9.0-beta.0) (2018-09-20)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.8.0...0.9.0-beta.0)

**Fixed bugs:**

- Don't add admission group as a prioritized version [\#156](https://github.com/kubedb/postgres/pull/156) ([tamalsaha](https://github.com/tamalsaha))

**Merged pull requests:**

- Pass extra args to tools.sh [\#196](https://github.com/kubedb/postgres/pull/196) ([the-redback](https://github.com/the-redback))
- Support Termination Policy & Stop working for deprecated \*Versions [\#195](https://github.com/kubedb/postgres/pull/195) ([hossainemruz](https://github.com/hossainemruz))
- Introduce synchronous streaming replication model [\#194](https://github.com/kubedb/postgres/pull/194) ([zhenhuadong](https://github.com/zhenhuadong))
- Revendor k8s.io/apiserver [\#193](https://github.com/kubedb/postgres/pull/193) ([tamalsaha](https://github.com/tamalsaha))
- Revendor kubernetes-1.11.3 [\#192](https://github.com/kubedb/postgres/pull/192) ([tamalsaha](https://github.com/tamalsaha))
- Support UpdateStrategy [\#190](https://github.com/kubedb/postgres/pull/190) ([tamalsaha](https://github.com/tamalsaha))
- Add TerminationPolicy for databases [\#189](https://github.com/kubedb/postgres/pull/189) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#188](https://github.com/kubedb/postgres/pull/188) ([tamalsaha](https://github.com/tamalsaha))
- Use IntHash as status.observedGeneration [\#187](https://github.com/kubedb/postgres/pull/187) ([tamalsaha](https://github.com/tamalsaha))
- fix build image [\#186](https://github.com/kubedb/postgres/pull/186) ([tahsinrahman](https://github.com/tahsinrahman))
- fix github status [\#185](https://github.com/kubedb/postgres/pull/185) ([tahsinrahman](https://github.com/tahsinrahman))
- update pipeline [\#184](https://github.com/kubedb/postgres/pull/184) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix E2E test for minikube [\#183](https://github.com/kubedb/postgres/pull/183) ([the-redback](https://github.com/the-redback))
- update pipeline [\#182](https://github.com/kubedb/postgres/pull/182) ([tahsinrahman](https://github.com/tahsinrahman))
- Update exporter image in concourse test [\#181](https://github.com/kubedb/postgres/pull/181) ([hossainemruz](https://github.com/hossainemruz))
- Migrate Postgres [\#180](https://github.com/kubedb/postgres/pull/180) ([tamalsaha](https://github.com/tamalsaha))
- Use exporter directly [\#179](https://github.com/kubedb/postgres/pull/179) ([hossainemruz](https://github.com/hossainemruz))
- Update status.ObservedGeneration for failure phase [\#178](https://github.com/kubedb/postgres/pull/178) ([the-redback](https://github.com/the-redback))
- Keep track of ObservedGenerationHash [\#177](https://github.com/kubedb/postgres/pull/177) ([tamalsaha](https://github.com/tamalsaha))
- Use NewObservableHandler [\#176](https://github.com/kubedb/postgres/pull/176) ([tamalsaha](https://github.com/tamalsaha))
- Fix uninstall for concourse [\#175](https://github.com/kubedb/postgres/pull/175) ([tahsinrahman](https://github.com/tahsinrahman))
- Revise verification of spec fields [\#174](https://github.com/kubedb/postgres/pull/174) ([tamalsaha](https://github.com/tamalsaha))
- Support passing args via PodTemplate [\#173](https://github.com/kubedb/postgres/pull/173) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#172](https://github.com/kubedb/postgres/pull/172) ([tamalsaha](https://github.com/tamalsaha))
- Introduce storageType : ephemeral [\#171](https://github.com/kubedb/postgres/pull/171) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#170](https://github.com/kubedb/postgres/pull/170) ([tamalsaha](https://github.com/tamalsaha))
- Add support for running tests on cncf machines [\#169](https://github.com/kubedb/postgres/pull/169) ([tahsinrahman](https://github.com/tahsinrahman))
- Keep track of observedGeneration in status [\#168](https://github.com/kubedb/postgres/pull/168) ([tamalsaha](https://github.com/tamalsaha))
- Use db crd image pull secrets as backup for snapshot jobs [\#167](https://github.com/kubedb/postgres/pull/167) ([tamalsaha](https://github.com/tamalsaha))
- Separate StatsService for monitoring [\#166](https://github.com/kubedb/postgres/pull/166) ([the-redback](https://github.com/the-redback))
- Use updated crd spec [\#165](https://github.com/kubedb/postgres/pull/165) ([tamalsaha](https://github.com/tamalsaha))
- Rename OffshootLabels to OffshootSelectors [\#164](https://github.com/kubedb/postgres/pull/164) ([tamalsaha](https://github.com/tamalsaha))
- Revendor apimachinery [\#163](https://github.com/kubedb/postgres/pull/163) ([tamalsaha](https://github.com/tamalsaha))
- Use PostgresVersion for postgres images [\#162](https://github.com/kubedb/postgres/pull/162) ([annymsMthd](https://github.com/annymsMthd))
- Revendor api [\#161](https://github.com/kubedb/postgres/pull/161) ([tamalsaha](https://github.com/tamalsaha))
- Use kmodules monitoring and objectstore api [\#160](https://github.com/kubedb/postgres/pull/160) ([tamalsaha](https://github.com/tamalsaha))
- Refactor concourse scripts [\#159](https://github.com/kubedb/postgres/pull/159) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix command `./hack/make.py test e2e` [\#158](https://github.com/kubedb/postgres/pull/158) ([the-redback](https://github.com/the-redback))
- Set generated binary name to pg-operator [\#157](https://github.com/kubedb/postgres/pull/157) ([tamalsaha](https://github.com/tamalsaha))
- Enable status subresource for crds [\#155](https://github.com/kubedb/postgres/pull/155) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v8.0.0 [\#154](https://github.com/kubedb/postgres/pull/154) ([tamalsaha](https://github.com/tamalsaha))
- Format shell scripts [\#153](https://github.com/kubedb/postgres/pull/153) ([tamalsaha](https://github.com/tamalsaha))
- Support custom configuration [\#152](https://github.com/kubedb/postgres/pull/152) ([hossainemruz](https://github.com/hossainemruz))
- Fix travis build [\#151](https://github.com/kubedb/postgres/pull/151) ([hossainemruz](https://github.com/hossainemruz))
- Updated osm version to 0.7.1 [\#150](https://github.com/kubedb/postgres/pull/150) ([the-redback](https://github.com/the-redback))
- Support ENV variables in CRDs [\#149](https://github.com/kubedb/postgres/pull/149) ([hossainemruz](https://github.com/hossainemruz))

## [0.8.0](https://github.com/kubedb/postgres/tree/0.8.0) (2018-06-12)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.8.0-rc.0...0.8.0)

**Merged pull requests:**

- Fix missing Error return [\#147](https://github.com/kubedb/postgres/pull/147) ([the-redback](https://github.com/the-redback))
- Revendor forked k8s.io/apiserver [\#146](https://github.com/kubedb/postgres/pull/146) ([tamalsaha](https://github.com/tamalsaha))
- Revendor dependencies for aws ans azure sdk [\#145](https://github.com/kubedb/postgres/pull/145) ([tamalsaha](https://github.com/tamalsaha))
- Add changelog [\#144](https://github.com/kubedb/postgres/pull/144) ([tamalsaha](https://github.com/tamalsaha))

## [0.8.0-rc.0](https://github.com/kubedb/postgres/tree/0.8.0-rc.0) (2018-05-28)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.8.0-beta.2...0.8.0-rc.0)

**Merged pull requests:**

- Update release script [\#143](https://github.com/kubedb/postgres/pull/143) ([tamalsaha](https://github.com/tamalsaha))
- Fixed kubeconfig plugin for Cloud Providers && Storage is required for Postgres [\#142](https://github.com/kubedb/postgres/pull/142) ([the-redback](https://github.com/the-redback))
-  Concourse [\#141](https://github.com/kubedb/postgres/pull/141) ([tahsinrahman](https://github.com/tahsinrahman))
-  Refactored E2E testing to support self-hosted operator with proper deployment configuration [\#140](https://github.com/kubedb/postgres/pull/140) ([the-redback](https://github.com/the-redback))
- Skip delete requests for empty resources [\#139](https://github.com/kubedb/postgres/pull/139) ([the-redback](https://github.com/the-redback))
- Don't panic if admission options is nil [\#138](https://github.com/kubedb/postgres/pull/138) ([tamalsaha](https://github.com/tamalsaha))
- Disable admission controllers for webhook server [\#137](https://github.com/kubedb/postgres/pull/137) ([tamalsaha](https://github.com/tamalsaha))
- Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 [\#136](https://github.com/kubedb/postgres/pull/136) ([the-redback](https://github.com/the-redback))
- Update client-go to 7.0.0 [\#135](https://github.com/kubedb/postgres/pull/135) ([tamalsaha](https://github.com/tamalsaha))
- Bundle Webhook Server and Added sharedinfomer Factory [\#132](https://github.com/kubedb/postgres/pull/132) ([the-redback](https://github.com/the-redback))
-  Moved ValidatingWebhook Packages from kubedb-server to postgres repo [\#131](https://github.com/kubedb/postgres/pull/131) ([the-redback](https://github.com/the-redback))
- Add travis yaml [\#130](https://github.com/kubedb/postgres/pull/130) ([tahsinrahman](https://github.com/tahsinrahman))

## [0.8.0-beta.2](https://github.com/kubedb/postgres/tree/0.8.0-beta.2) (2018-02-27)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.8.0-beta.1...0.8.0-beta.2)

**Implemented enhancements:**

- use separate script for different task [\#126](https://github.com/kubedb/postgres/pull/126) ([aerokite](https://github.com/aerokite))

**Fixed bugs:**

- use separate script for different task [\#126](https://github.com/kubedb/postgres/pull/126) ([aerokite](https://github.com/aerokite))

**Merged pull requests:**

- Use apps/v1 [\#128](https://github.com/kubedb/postgres/pull/128) ([aerokite](https://github.com/aerokite))
- upgrade version & fixed service [\#127](https://github.com/kubedb/postgres/pull/127) ([aerokite](https://github.com/aerokite))
- Fix for pointer type [\#125](https://github.com/kubedb/postgres/pull/125) ([aerokite](https://github.com/aerokite))
- Fix dormantDB matching: pass same type to Equal method [\#124](https://github.com/kubedb/postgres/pull/124) ([the-redback](https://github.com/the-redback))
- Add support of Postgres 10.2 [\#123](https://github.com/kubedb/postgres/pull/123) ([aerokite](https://github.com/aerokite))
- Fixed dormantdb matching & Raised throttling time & Fixed Postgres version checking [\#121](https://github.com/kubedb/postgres/pull/121) ([the-redback](https://github.com/the-redback))
- Use official code generator scripts [\#120](https://github.com/kubedb/postgres/pull/120) ([tamalsaha](https://github.com/tamalsaha))
- Fix merge service ports [\#119](https://github.com/kubedb/postgres/pull/119) ([aerokite](https://github.com/aerokite))

## [0.8.0-beta.1](https://github.com/kubedb/postgres/tree/0.8.0-beta.1) (2018-01-29)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.8.0-beta.0...0.8.0-beta.1)

**Merged pull requests:**

- Reorg docker code structure [\#117](https://github.com/kubedb/postgres/pull/117) ([aerokite](https://github.com/aerokite))

## [0.8.0-beta.0](https://github.com/kubedb/postgres/tree/0.8.0-beta.0) (2018-01-07)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.7.1...0.8.0-beta.0)

**Merged pull requests:**

- Update rbac role [\#116](https://github.com/kubedb/postgres/pull/116) ([aerokite](https://github.com/aerokite))
- Use work queue [\#114](https://github.com/kubedb/postgres/pull/114) ([aerokite](https://github.com/aerokite))
- Set client id for analytics [\#112](https://github.com/kubedb/postgres/pull/112) ([tamalsaha](https://github.com/tamalsaha))
- Fix CRD registration [\#107](https://github.com/kubedb/postgres/pull/107) ([the-redback](https://github.com/the-redback))
- Added log-based archive support with wal-g in postgres [\#106](https://github.com/kubedb/postgres/pull/106) ([aerokite](https://github.com/aerokite))
- Remove dependency on deleted appscode/log packages. [\#105](https://github.com/kubedb/postgres/pull/105) ([tamalsaha](https://github.com/tamalsaha))
- Use monitoring tools from appscode/kutil [\#104](https://github.com/kubedb/postgres/pull/104) ([tamalsaha](https://github.com/tamalsaha))
- fixes k8sdb/operator\#126 for postgres part [\#103](https://github.com/kubedb/postgres/pull/103) ([the-redback](https://github.com/the-redback))
- Use client-go 5.x [\#102](https://github.com/kubedb/postgres/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Update secret procedure for Restore [\#101](https://github.com/kubedb/postgres/pull/101) ([the-redback](https://github.com/the-redback))

## [0.7.1](https://github.com/kubedb/postgres/tree/0.7.1) (2017-10-04)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.7.0...0.7.1)

## [0.7.0](https://github.com/kubedb/postgres/tree/0.7.0) (2017-09-26)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.6.0...0.7.0)

**Merged pull requests:**

- Assign Kind Type in CRD object [\#100](https://github.com/kubedb/postgres/pull/100) ([aerokite](https://github.com/aerokite))
- Set Affinity and Tolerations from CRD spec [\#99](https://github.com/kubedb/postgres/pull/99) ([tamalsaha](https://github.com/tamalsaha))
- Support migration from TPR to CRD [\#98](https://github.com/kubedb/postgres/pull/98) ([aerokite](https://github.com/aerokite))
- Use kutil in e2e-test [\#97](https://github.com/kubedb/postgres/pull/97) ([aerokite](https://github.com/aerokite))
- Resume DormantDatabase while creating Original DB again [\#96](https://github.com/kubedb/postgres/pull/96) ([aerokite](https://github.com/aerokite))
- Rewrite e2e tests using ginkgo [\#95](https://github.com/kubedb/postgres/pull/95) ([aerokite](https://github.com/aerokite))

## [0.6.0](https://github.com/kubedb/postgres/tree/0.6.0) (2017-07-24)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.5.0...0.6.0)

**Merged pull requests:**

- Revendor for api fix [\#94](https://github.com/kubedb/postgres/pull/94) ([aerokite](https://github.com/aerokite))

## [0.5.0](https://github.com/kubedb/postgres/tree/0.5.0) (2017-07-19)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.4.0...0.5.0)

## [0.4.0](https://github.com/kubedb/postgres/tree/0.4.0) (2017-07-18)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.3.1...0.4.0)

## [0.3.1](https://github.com/kubedb/postgres/tree/0.3.1) (2017-07-14)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.3.0...0.3.1)

## [0.3.0](https://github.com/kubedb/postgres/tree/0.3.0) (2017-07-08)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.2.0...0.3.0)

**Merged pull requests:**

- Support RBAC [\#92](https://github.com/kubedb/postgres/pull/92) ([aerokite](https://github.com/aerokite))
- Allow setting resources for StatefulSet or Snapshot/Restore jobs [\#91](https://github.com/kubedb/postgres/pull/91) ([tamalsaha](https://github.com/tamalsaha))
- Use updated snapshot storage format [\#90](https://github.com/kubedb/postgres/pull/90) ([tamalsaha](https://github.com/tamalsaha))
- Add app=kubedb labels to TPR reg [\#89](https://github.com/kubedb/postgres/pull/89) ([tamalsaha](https://github.com/tamalsaha))
- Support using non-default service account [\#88](https://github.com/kubedb/postgres/pull/88) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0](https://github.com/kubedb/postgres/tree/0.2.0) (2017-06-22)
[Full Changelog](https://github.com/kubedb/postgres/compare/0.1.0...0.2.0)

**Merged pull requests:**

- Expose exporter port via service [\#86](https://github.com/kubedb/postgres/pull/86) ([tamalsaha](https://github.com/tamalsaha))
- Correctly parse target port [\#85](https://github.com/kubedb/postgres/pull/85) ([tamalsaha](https://github.com/tamalsaha))
- Run side car exporter [\#84](https://github.com/kubedb/postgres/pull/84) ([tamalsaha](https://github.com/tamalsaha))
- get summary report [\#83](https://github.com/kubedb/postgres/pull/83) ([aerokite](https://github.com/aerokite))
- Use client-go [\#82](https://github.com/kubedb/postgres/pull/82) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0](https://github.com/kubedb/postgres/tree/0.1.0) (2017-06-14)
**Fixed bugs:**

- Allow updating to create missing workloads [\#78](https://github.com/kubedb/postgres/pull/78) ([aerokite](https://github.com/aerokite))

**Merged pull requests:**

- Change api version to v1alpha1 [\#81](https://github.com/kubedb/postgres/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- Pass cronController as parameter [\#80](https://github.com/kubedb/postgres/pull/80) ([aerokite](https://github.com/aerokite))
- Use built-in exporter [\#79](https://github.com/kubedb/postgres/pull/79) ([tamalsaha](https://github.com/tamalsaha))
- Add analytics event for operator [\#77](https://github.com/kubedb/postgres/pull/77) ([aerokite](https://github.com/aerokite))
- Add analytics [\#76](https://github.com/kubedb/postgres/pull/76) ([aerokite](https://github.com/aerokite))
- Use util tag matching TPR version [\#75](https://github.com/kubedb/postgres/pull/75) ([tamalsaha](https://github.com/tamalsaha))
- Revendor client-go [\#74](https://github.com/kubedb/postgres/pull/74) ([tamalsaha](https://github.com/tamalsaha))
- Add Run\(\) method to just run controller. [\#73](https://github.com/kubedb/postgres/pull/73) ([tamalsaha](https://github.com/tamalsaha))
- Add HTTP server to expose metrics [\#72](https://github.com/kubedb/postgres/pull/72) ([tamalsaha](https://github.com/tamalsaha))
- Prometheus support [\#71](https://github.com/kubedb/postgres/pull/71) ([saumanbiswas](https://github.com/saumanbiswas))
- Use kubedb docker hub account [\#70](https://github.com/kubedb/postgres/pull/70) ([tamalsaha](https://github.com/tamalsaha))
- Rename operator name [\#69](https://github.com/kubedb/postgres/pull/69) ([aerokite](https://github.com/aerokite))
- Use kubedb.com apigroup instead of k8sdb.com [\#68](https://github.com/kubedb/postgres/pull/68) ([tamalsaha](https://github.com/tamalsaha))
- Pass clients instead of config [\#66](https://github.com/kubedb/postgres/pull/66) ([aerokite](https://github.com/aerokite))
- Ungroup imports on fmt [\#65](https://github.com/kubedb/postgres/pull/65) ([tamalsaha](https://github.com/tamalsaha))
- Fix go report card issues [\#64](https://github.com/kubedb/postgres/pull/64) ([tamalsaha](https://github.com/tamalsaha))
- Use common receiver [\#63](https://github.com/kubedb/postgres/pull/63) ([tamalsaha](https://github.com/tamalsaha))
- Rename delete database to pause [\#62](https://github.com/kubedb/postgres/pull/62) ([tamalsaha](https://github.com/tamalsaha))
- Rename DeletedDatabase to DormantDatabase [\#61](https://github.com/kubedb/postgres/pull/61) ([tamalsaha](https://github.com/tamalsaha))
- Fix update method [\#59](https://github.com/kubedb/postgres/pull/59) ([aerokite](https://github.com/aerokite))
- Remove prefix from snapshot job [\#58](https://github.com/kubedb/postgres/pull/58) ([aerokite](https://github.com/aerokite))
- Delete Database Secret for wipe out [\#57](https://github.com/kubedb/postgres/pull/57) ([aerokite](https://github.com/aerokite))
- Rename DatabaseSnapshot to Snapshot [\#56](https://github.com/kubedb/postgres/pull/56) ([tamalsaha](https://github.com/tamalsaha))
- Modify StatefulSet naming format [\#54](https://github.com/kubedb/postgres/pull/54) ([aerokite](https://github.com/aerokite))
- Get object each time before updating [\#53](https://github.com/kubedb/postgres/pull/53) ([aerokite](https://github.com/aerokite))
- Create headless service for StatefulSet [\#51](https://github.com/kubedb/postgres/pull/51) ([aerokite](https://github.com/aerokite))
- Use data as Volume name [\#50](https://github.com/kubedb/postgres/pull/50) ([aerokite](https://github.com/aerokite))
- Put kind in label instead of type [\#48](https://github.com/kubedb/postgres/pull/48) ([aerokite](https://github.com/aerokite))
- Do not store autogenerated meta information [\#47](https://github.com/kubedb/postgres/pull/47) ([aerokite](https://github.com/aerokite))
- Bubble up error for controller methods [\#45](https://github.com/kubedb/postgres/pull/45) ([aerokite](https://github.com/aerokite))
- Modify e2e test. Do not support recovery by recreating Postgres anymore [\#44](https://github.com/kubedb/postgres/pull/44) ([aerokite](https://github.com/aerokite))
- Use Kubernetes EventRecorder directly [\#43](https://github.com/kubedb/postgres/pull/43) ([aerokite](https://github.com/aerokite))
- Address status field changes [\#42](https://github.com/kubedb/postgres/pull/42) ([aerokite](https://github.com/aerokite))
- Use canary tag for k8sdb images [\#40](https://github.com/kubedb/postgres/pull/40) ([aerokite](https://github.com/aerokite))
- Install ca-certificates in operator docker image. [\#39](https://github.com/kubedb/postgres/pull/39) ([tamalsaha](https://github.com/tamalsaha))
- Add deployment.yaml [\#38](https://github.com/kubedb/postgres/pull/38) ([aerokite](https://github.com/aerokite))
- Rename "destroy" to "wipeOut" [\#36](https://github.com/kubedb/postgres/pull/36) ([tamalsaha](https://github.com/tamalsaha))
- Store Postgres Spec in DeletedDatabase [\#34](https://github.com/kubedb/postgres/pull/34) ([aerokite](https://github.com/aerokite))
- Update timing fields [\#33](https://github.com/kubedb/postgres/pull/33) ([tamalsaha](https://github.com/tamalsaha))
- Remove -v\* suffix from docker image [\#32](https://github.com/kubedb/postgres/pull/32) ([tamalsaha](https://github.com/tamalsaha))
- Use k8sdb docker hub account [\#31](https://github.com/kubedb/postgres/pull/31) ([tamalsaha](https://github.com/tamalsaha))
- Support initialization using DatabaseSnapshot [\#30](https://github.com/kubedb/postgres/pull/30) ([aerokite](https://github.com/aerokite))
- Use resource name constant from apimachinery [\#29](https://github.com/kubedb/postgres/pull/29) ([tamalsaha](https://github.com/tamalsaha))
- Use one controller struct [\#28](https://github.com/kubedb/postgres/pull/28) ([tamalsaha](https://github.com/tamalsaha))
- Implement updated interfaces. [\#27](https://github.com/kubedb/postgres/pull/27) ([tamalsaha](https://github.com/tamalsaha))
- Rename controller image to k8s-pg [\#26](https://github.com/kubedb/postgres/pull/26) ([tamalsaha](https://github.com/tamalsaha))
- Implement Snapshotter, Deleter with Controller [\#25](https://github.com/kubedb/postgres/pull/25) ([aerokite](https://github.com/aerokite))
- Implement recover operation [\#24](https://github.com/kubedb/postgres/pull/24) ([aerokite](https://github.com/aerokite))
- Implement k8sdb framework [\#23](https://github.com/kubedb/postgres/pull/23) ([aerokite](https://github.com/aerokite))
- Use osm to pull/push snapshots [\#22](https://github.com/kubedb/postgres/pull/22) ([aerokite](https://github.com/aerokite))
- Modify [\#19](https://github.com/kubedb/postgres/pull/19) ([aerokite](https://github.com/aerokite))
- Fix [\#18](https://github.com/kubedb/postgres/pull/18) ([aerokite](https://github.com/aerokite))
- Remove "volume.alpha.kubernetes.io/storage-class" annotation [\#14](https://github.com/kubedb/postgres/pull/14) ([aerokite](https://github.com/aerokite))
- add controller operation & docker files [\#2](https://github.com/kubedb/postgres/pull/2) ([aerokite](https://github.com/aerokite))
- Modify skeleton to postgres [\#1](https://github.com/kubedb/postgres/pull/1) ([aerokite](https://github.com/aerokite))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*