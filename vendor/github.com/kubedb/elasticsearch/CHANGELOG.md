# Change Log

## [0.9.0-rc.0](https://github.com/kubedb/elasticsearch/tree/0.9.0-rc.0) (2018-10-15)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.9.0-beta.1...0.9.0-rc.0)

**Merged pull requests:**

- Support providing resources for monitoring container [\#223](https://github.com/kubedb/elasticsearch/pull/223) ([hossainemruz](https://github.com/hossainemruz))
- Recognize denied request by any webhook in xray [\#222](https://github.com/kubedb/elasticsearch/pull/222) ([tamalsaha](https://github.com/tamalsaha))
- Update kubernetes client libraries to 1.12.0 [\#221](https://github.com/kubedb/elasticsearch/pull/221) ([tamalsaha](https://github.com/tamalsaha))
- Add validation webhook xray [\#220](https://github.com/kubedb/elasticsearch/pull/220) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes [\#219](https://github.com/kubedb/elasticsearch/pull/219) ([hossainemruz](https://github.com/hossainemruz))
- Update  tools image for different Auth Plugin [\#218](https://github.com/kubedb/elasticsearch/pull/218) ([hossainemruz](https://github.com/hossainemruz))
- Fix storage validation [\#217](https://github.com/kubedb/elasticsearch/pull/217) ([hossainemruz](https://github.com/hossainemruz))
- Merge ports from service template [\#215](https://github.com/kubedb/elasticsearch/pull/215) ([tamalsaha](https://github.com/tamalsaha))
- Replace doNotPause with TerminationPolicy = DoNotTerminate [\#214](https://github.com/kubedb/elasticsearch/pull/214) ([tamalsaha](https://github.com/tamalsaha))
- Disable Search Guard & Support Elasticsearch 6.4 [\#213](https://github.com/kubedb/elasticsearch/pull/213) ([hossainemruz](https://github.com/hossainemruz))
- Pass resources to NamespaceValidator [\#212](https://github.com/kubedb/elasticsearch/pull/212) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes [\#211](https://github.com/kubedb/elasticsearch/pull/211) ([tamalsaha](https://github.com/tamalsaha))
- Support Livecycle hook and container probes [\#210](https://github.com/kubedb/elasticsearch/pull/210) ([tamalsaha](https://github.com/tamalsaha))
- Check if Kubernetes version is supported before running operator [\#209](https://github.com/kubedb/elasticsearch/pull/209) ([tamalsaha](https://github.com/tamalsaha))
- Update package alias [\#208](https://github.com/kubedb/elasticsearch/pull/208) ([tamalsaha](https://github.com/tamalsaha))

## [0.9.0-beta.1](https://github.com/kubedb/elasticsearch/tree/0.9.0-beta.1) (2018-09-30)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.9.0-beta.0...0.9.0-beta.1)

**Merged pull requests:**

- Revendor api [\#207](https://github.com/kubedb/elasticsearch/pull/207) ([tamalsaha](https://github.com/tamalsaha))
- Use spec.authPlugin for Elasticsearch [\#206](https://github.com/kubedb/elasticsearch/pull/206) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api for catalog apigroup [\#205](https://github.com/kubedb/elasticsearch/pull/205) ([tamalsaha](https://github.com/tamalsaha))
- Use resources from podTemplate.spec.resources [\#204](https://github.com/kubedb/elasticsearch/pull/204) ([hossainemruz](https://github.com/hossainemruz))
- Use --pull flag with docker build \(\#20\) [\#203](https://github.com/kubedb/elasticsearch/pull/203) ([tamalsaha](https://github.com/tamalsaha))

## [0.9.0-beta.0](https://github.com/kubedb/elasticsearch/tree/0.9.0-beta.0) (2018-09-20)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.8.0...0.9.0-beta.0)

**Fixed bugs:**

- Search used secrets within same namespace of ES [\#200](https://github.com/kubedb/elasticsearch/pull/200) ([tamalsaha](https://github.com/tamalsaha))

**Merged pull requests:**

- Show Deprecated column for Elasticsearchversions [\#202](https://github.com/kubedb/elasticsearch/pull/202) ([hossainemruz](https://github.com/hossainemruz))
- Pass extra args to tools.sh [\#201](https://github.com/kubedb/elasticsearch/pull/201) ([the-redback](https://github.com/the-redback))
- Use suffix for updated DBImage & Stop working for deprecated \*Versions [\#199](https://github.com/kubedb/elasticsearch/pull/199) ([the-redback](https://github.com/the-redback))
- Don't try to wipe out Snapshot data for Local backend [\#198](https://github.com/kubedb/elasticsearch/pull/198) ([hossainemruz](https://github.com/hossainemruz))
- Revendor k8s.io/apiserver [\#197](https://github.com/kubedb/elasticsearch/pull/197) ([tamalsaha](https://github.com/tamalsaha))
- Revendor kubernetes-1.11.3 [\#196](https://github.com/kubedb/elasticsearch/pull/196) ([tamalsaha](https://github.com/tamalsaha))
- Support UpdateStrategy [\#195](https://github.com/kubedb/elasticsearch/pull/195) ([tamalsaha](https://github.com/tamalsaha))
-  Support Termination Policy [\#194](https://github.com/kubedb/elasticsearch/pull/194) ([the-redback](https://github.com/the-redback))
- Add TerminationPolicy for databases [\#193](https://github.com/kubedb/elasticsearch/pull/193) ([tamalsaha](https://github.com/tamalsaha))
- Fix http/https scheme for EnableSSL [\#192](https://github.com/kubedb/elasticsearch/pull/192) ([the-redback](https://github.com/the-redback))
- Revendor api [\#191](https://github.com/kubedb/elasticsearch/pull/191) ([tamalsaha](https://github.com/tamalsaha))
- Fix build [\#190](https://github.com/kubedb/elasticsearch/pull/190) ([tamalsaha](https://github.com/tamalsaha))
- Use IntHash in AlreadyObserved helpers [\#189](https://github.com/kubedb/elasticsearch/pull/189) ([tamalsaha](https://github.com/tamalsaha))
- Improve error message for GetIndices [\#188](https://github.com/kubedb/elasticsearch/pull/188) ([the-redback](https://github.com/the-redback))
- fix github status [\#187](https://github.com/kubedb/elasticsearch/pull/187) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix E2E test for minikube & Fixed elasticsearch db image 5.6.4 [\#186](https://github.com/kubedb/elasticsearch/pull/186) ([the-redback](https://github.com/the-redback))
- update pipeline [\#185](https://github.com/kubedb/elasticsearch/pull/185) ([tahsinrahman](https://github.com/tahsinrahman))
- Add Kibana docker files [\#184](https://github.com/kubedb/elasticsearch/pull/184) ([hossainemruz](https://github.com/hossainemruz))
- update pipeline [\#183](https://github.com/kubedb/elasticsearch/pull/183) ([tahsinrahman](https://github.com/tahsinrahman))
- Disable Search Guard enterprise modules by default [\#182](https://github.com/kubedb/elasticsearch/pull/182) ([hossainemruz](https://github.com/hossainemruz))
- Use Exporters directly [\#181](https://github.com/kubedb/elasticsearch/pull/181) ([hossainemruz](https://github.com/hossainemruz))
- Migrate elasticsearch [\#180](https://github.com/kubedb/elasticsearch/pull/180) ([tamalsaha](https://github.com/tamalsaha))
- Update status.ObservedGeneration for failure phase [\#179](https://github.com/kubedb/elasticsearch/pull/179) ([the-redback](https://github.com/the-redback))
- Set `USE\_SSL` env in exporter container [\#178](https://github.com/kubedb/elasticsearch/pull/178) ([hossainemruz](https://github.com/hossainemruz))
- Use NewObservableHandler [\#177](https://github.com/kubedb/elasticsearch/pull/177) ([tamalsaha](https://github.com/tamalsaha))
- Fix uninstall for concourse [\#176](https://github.com/kubedb/elasticsearch/pull/176) ([tahsinrahman](https://github.com/tahsinrahman))
- Use Search Guard 23.0 for Elasticsearch 6.3 [\#175](https://github.com/kubedb/elasticsearch/pull/175) ([hossainemruz](https://github.com/hossainemruz))
- Revised verification of spec fields [\#174](https://github.com/kubedb/elasticsearch/pull/174) ([the-redback](https://github.com/the-redback))
- Support passing args via PodTemplate [\#173](https://github.com/kubedb/elasticsearch/pull/173) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#172](https://github.com/kubedb/elasticsearch/pull/172) ([tamalsaha](https://github.com/tamalsaha))
- Update error message [\#171](https://github.com/kubedb/elasticsearch/pull/171) ([tamalsaha](https://github.com/tamalsaha))
- Introduce storageType : ephemeral [\#170](https://github.com/kubedb/elasticsearch/pull/170) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#169](https://github.com/kubedb/elasticsearch/pull/169) ([tamalsaha](https://github.com/tamalsaha))
- Add support for running tests on cncf cluster [\#168](https://github.com/kubedb/elasticsearch/pull/168) ([tahsinrahman](https://github.com/tahsinrahman))
- Support both .yml and .yaml for config files [\#167](https://github.com/kubedb/elasticsearch/pull/167) ([hossainemruz](https://github.com/hossainemruz))
- Keep track of observedGeneration in status [\#166](https://github.com/kubedb/elasticsearch/pull/166) ([tamalsaha](https://github.com/tamalsaha))
- Fix comments for stats service [\#165](https://github.com/kubedb/elasticsearch/pull/165) ([tamalsaha](https://github.com/tamalsaha))
-  Separate StatsService for monitoring [\#164](https://github.com/kubedb/elasticsearch/pull/164) ([the-redback](https://github.com/the-redback))
-  Use ElasticsearchVersion for Elasticsearch images [\#163](https://github.com/kubedb/elasticsearch/pull/163) ([the-redback](https://github.com/the-redback))
- Use updated crd spec [\#162](https://github.com/kubedb/elasticsearch/pull/162) ([tamalsaha](https://github.com/tamalsaha))
- Rename OffshootLabels to OffshootSelectors [\#161](https://github.com/kubedb/elasticsearch/pull/161) ([tamalsaha](https://github.com/tamalsaha))
- Revendor api [\#160](https://github.com/kubedb/elasticsearch/pull/160) ([tamalsaha](https://github.com/tamalsaha))
- Fix docker for es-6.3.0 and revendored [\#159](https://github.com/kubedb/elasticsearch/pull/159) ([the-redback](https://github.com/the-redback))
- Use kmodules monitoring and objectstore api [\#158](https://github.com/kubedb/elasticsearch/pull/158) ([tamalsaha](https://github.com/tamalsaha))
- Support Elasticsearch 6.3 [\#157](https://github.com/kubedb/elasticsearch/pull/157) ([tamalsaha](https://github.com/tamalsaha))
- Support custom configuration  [\#156](https://github.com/kubedb/elasticsearch/pull/156) ([hossainemruz](https://github.com/hossainemruz))
- Refactor concourse scripts [\#155](https://github.com/kubedb/elasticsearch/pull/155) ([tahsinrahman](https://github.com/tahsinrahman))
- Fix command `./hack/make.py test e2e` [\#154](https://github.com/kubedb/elasticsearch/pull/154) ([the-redback](https://github.com/the-redback))
- Set generated binary name to es-operator [\#153](https://github.com/kubedb/elasticsearch/pull/153) ([tamalsaha](https://github.com/tamalsaha))
- Don't add admission/v1beta1 group as a prioritized version [\#152](https://github.com/kubedb/elasticsearch/pull/152) ([tamalsaha](https://github.com/tamalsaha))
- Enable status subresource for crds [\#151](https://github.com/kubedb/elasticsearch/pull/151) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to v8.0.0 [\#150](https://github.com/kubedb/elasticsearch/pull/150) ([tamalsaha](https://github.com/tamalsaha))
- Format shell script [\#149](https://github.com/kubedb/elasticsearch/pull/149) ([tamalsaha](https://github.com/tamalsaha))
- Support ENV variables in CRDs [\#146](https://github.com/kubedb/elasticsearch/pull/146) ([hossainemruz](https://github.com/hossainemruz))
- Updated osm version to 0.7.1 [\#145](https://github.com/kubedb/elasticsearch/pull/145) ([the-redback](https://github.com/the-redback))

## [0.8.0](https://github.com/kubedb/elasticsearch/tree/0.8.0) (2018-06-12)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.8.0-rc.0...0.8.0)

**Merged pull requests:**

- Fix missing error return [\#144](https://github.com/kubedb/elasticsearch/pull/144) ([the-redback](https://github.com/the-redback))
-  Tests and operator changed to support es-6.2.4 [\#143](https://github.com/kubedb/elasticsearch/pull/143) ([the-redback](https://github.com/the-redback))
- Added ES 6.2.4 Docker build files [\#142](https://github.com/kubedb/elasticsearch/pull/142) ([stormmore](https://github.com/stormmore))
- Revendor dependencies [\#141](https://github.com/kubedb/elasticsearch/pull/141) ([tamalsaha](https://github.com/tamalsaha))
- Support disabling Search Guard [\#140](https://github.com/kubedb/elasticsearch/pull/140) ([tamalsaha](https://github.com/tamalsaha))
- Add changelog [\#139](https://github.com/kubedb/elasticsearch/pull/139) ([tamalsaha](https://github.com/tamalsaha))

## [0.8.0-rc.0](https://github.com/kubedb/elasticsearch/tree/0.8.0-rc.0) (2018-05-28)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.8.0-beta.2...0.8.0-rc.0)

**Merged pull requests:**

-  Initialize database heapsize from Resource.requests [\#138](https://github.com/kubedb/elasticsearch/pull/138) ([the-redback](https://github.com/the-redback))
- concourse [\#137](https://github.com/kubedb/elasticsearch/pull/137) ([tahsinrahman](https://github.com/tahsinrahman))
- Refactored E2E testing to support self-hosted operator with proper deployment configuration [\#136](https://github.com/kubedb/elasticsearch/pull/136) ([the-redback](https://github.com/the-redback))
- to allow Request header field Authorization [\#135](https://github.com/kubedb/elasticsearch/pull/135) ([aerokite](https://github.com/aerokite))
- Skip delete requests for empty resources [\#134](https://github.com/kubedb/elasticsearch/pull/134) ([the-redback](https://github.com/the-redback))
- Use separate resource & storage [\#133](https://github.com/kubedb/elasticsearch/pull/133) ([aerokite](https://github.com/aerokite))
- Don't panic if admission options is nil [\#132](https://github.com/kubedb/elasticsearch/pull/132) ([tamalsaha](https://github.com/tamalsaha))
- Disable admission controllers for webhook server [\#131](https://github.com/kubedb/elasticsearch/pull/131) ([tamalsaha](https://github.com/tamalsaha))
- Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 [\#130](https://github.com/kubedb/elasticsearch/pull/130) ([the-redback](https://github.com/the-redback))
- Update client-go to 7.0.0 [\#129](https://github.com/kubedb/elasticsearch/pull/129) ([tamalsaha](https://github.com/tamalsaha))
- Bundle webhook server & Used  SharedInformer Factory with n-EventHandler [\#128](https://github.com/kubedb/elasticsearch/pull/128) ([the-redback](https://github.com/the-redback))
- Moved admission webhook packages into elasticsearch repo [\#127](https://github.com/kubedb/elasticsearch/pull/127) ([the-redback](https://github.com/the-redback))
- Add travis yaml [\#125](https://github.com/kubedb/elasticsearch/pull/125) ([tahsinrahman](https://github.com/tahsinrahman))

## [0.8.0-beta.2](https://github.com/kubedb/elasticsearch/tree/0.8.0-beta.2) (2018-02-27)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.8.0-beta.1...0.8.0-beta.2)

**Merged pull requests:**

- Use apps/v1 [\#123](https://github.com/kubedb/elasticsearch/pull/123) ([aerokite](https://github.com/aerokite))
- update validation [\#122](https://github.com/kubedb/elasticsearch/pull/122) ([aerokite](https://github.com/aerokite))
- Fix for pointer Type [\#121](https://github.com/kubedb/elasticsearch/pull/121) ([aerokite](https://github.com/aerokite))
- pass same type to Equal method [\#120](https://github.com/kubedb/elasticsearch/pull/120) ([aerokite](https://github.com/aerokite))
- Fixed dormantdb matching & Raised throttling time & Fixed Elasticsearch version checking [\#118](https://github.com/kubedb/elasticsearch/pull/118) ([the-redback](https://github.com/the-redback))
- Use official code generator scripts [\#117](https://github.com/kubedb/elasticsearch/pull/117) ([tamalsaha](https://github.com/tamalsaha))
- Use github.com/pkg/errors [\#116](https://github.com/kubedb/elasticsearch/pull/116) ([tamalsaha](https://github.com/tamalsaha))
- Use separate certs for node & client and use random password by default [\#115](https://github.com/kubedb/elasticsearch/pull/115) ([aerokite](https://github.com/aerokite))
- Fix pluralization of Elasticsearch [\#114](https://github.com/kubedb/elasticsearch/pull/114) ([tamalsaha](https://github.com/tamalsaha))

## [0.8.0-beta.1](https://github.com/kubedb/elasticsearch/tree/0.8.0-beta.1) (2018-01-29)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.8.0-beta.0...0.8.0-beta.1)

**Merged pull requests:**

- Fix for Job watcher [\#111](https://github.com/kubedb/elasticsearch/pull/111) ([aerokite](https://github.com/aerokite))
- reorg docker code structure [\#110](https://github.com/kubedb/elasticsearch/pull/110) ([aerokite](https://github.com/aerokite))

## [0.8.0-beta.0](https://github.com/kubedb/elasticsearch/tree/0.8.0-beta.0) (2018-01-07)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.7.1...0.8.0-beta.0)

**Merged pull requests:**

- pass analytics client-id as ENV [\#108](https://github.com/kubedb/elasticsearch/pull/108) ([aerokite](https://github.com/aerokite))
- Use work queue [\#106](https://github.com/kubedb/elasticsearch/pull/106) ([aerokite](https://github.com/aerokite))
- Reorg location of docker images [\#105](https://github.com/kubedb/elasticsearch/pull/105) ([aerokite](https://github.com/aerokite))
- Set client id for analytics [\#104](https://github.com/kubedb/elasticsearch/pull/104) ([tamalsaha](https://github.com/tamalsaha))
- Add explanation for oid bytes [\#103](https://github.com/kubedb/elasticsearch/pull/103) ([tamalsaha](https://github.com/tamalsaha))
- Revendor [\#102](https://github.com/kubedb/elasticsearch/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Fix CRD registration [\#100](https://github.com/kubedb/elasticsearch/pull/100) ([the-redback](https://github.com/the-redback))
- Remove deleted appcode/log package [\#99](https://github.com/kubedb/elasticsearch/pull/99) ([tamalsaha](https://github.com/tamalsaha))
- Use monitoring tools from appscode/kutil [\#98](https://github.com/kubedb/elasticsearch/pull/98) ([tamalsaha](https://github.com/tamalsaha))
- Support elasticsearch 5.6.3 with dedicated nodes [\#97](https://github.com/kubedb/elasticsearch/pull/97) ([aerokite](https://github.com/aerokite))
- Use client-go 5.x [\#96](https://github.com/kubedb/elasticsearch/pull/96) ([tamalsaha](https://github.com/tamalsaha))

## [0.7.1](https://github.com/kubedb/elasticsearch/tree/0.7.1) (2017-10-04)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.7.0...0.7.1)

## [0.7.0](https://github.com/kubedb/elasticsearch/tree/0.7.0) (2017-09-26)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.6.0...0.7.0)

**Merged pull requests:**

- Set Affinity and Tolerations from CRD spec. [\#95](https://github.com/kubedb/elasticsearch/pull/95) ([tamalsaha](https://github.com/tamalsaha))
- Support migration from TPR to CRD [\#94](https://github.com/kubedb/elasticsearch/pull/94) ([aerokite](https://github.com/aerokite))
- Use kutil in e2e-test [\#93](https://github.com/kubedb/elasticsearch/pull/93) ([aerokite](https://github.com/aerokite))
- Resume DormantDatabase while creating Original DB again [\#92](https://github.com/kubedb/elasticsearch/pull/92) ([aerokite](https://github.com/aerokite))
- Rewrite e2e tests using ginkgo [\#91](https://github.com/kubedb/elasticsearch/pull/91) ([aerokite](https://github.com/aerokite))

## [0.6.0](https://github.com/kubedb/elasticsearch/tree/0.6.0) (2017-07-24)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.5.0...0.6.0)

**Merged pull requests:**

- Revendor for api fix [\#90](https://github.com/kubedb/elasticsearch/pull/90) ([aerokite](https://github.com/aerokite))

## [0.5.0](https://github.com/kubedb/elasticsearch/tree/0.5.0) (2017-07-19)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.4.0...0.5.0)

## [0.4.0](https://github.com/kubedb/elasticsearch/tree/0.4.0) (2017-07-18)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.3.1...0.4.0)

## [0.3.1](https://github.com/kubedb/elasticsearch/tree/0.3.1) (2017-07-14)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.3.0...0.3.1)

## [0.3.0](https://github.com/kubedb/elasticsearch/tree/0.3.0) (2017-07-08)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.2.0...0.3.0)

**Merged pull requests:**

- 	Support RBAC [\#89](https://github.com/kubedb/elasticsearch/pull/89) ([aerokite](https://github.com/aerokite))
- Use snapshot path prefix  [\#88](https://github.com/kubedb/elasticsearch/pull/88) ([tamalsaha](https://github.com/tamalsaha))
- Allow setting resources for StatefulSet or Snapshot/Restore jobs [\#87](https://github.com/kubedb/elasticsearch/pull/87) ([tamalsaha](https://github.com/tamalsaha))
- Add app=kubedb label to TPR registration [\#86](https://github.com/kubedb/elasticsearch/pull/86) ([tamalsaha](https://github.com/tamalsaha))
- Support non-default service account with offshoot pods [\#85](https://github.com/kubedb/elasticsearch/pull/85) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0](https://github.com/kubedb/elasticsearch/tree/0.2.0) (2017-06-22)
[Full Changelog](https://github.com/kubedb/elasticsearch/compare/0.1.0...0.2.0)

**Merged pull requests:**

- Expose exporter port via service [\#83](https://github.com/kubedb/elasticsearch/pull/83) ([tamalsaha](https://github.com/tamalsaha))
- get summary report [\#82](https://github.com/kubedb/elasticsearch/pull/82) ([aerokite](https://github.com/aerokite))
- Use side-car exporter [\#81](https://github.com/kubedb/elasticsearch/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- Use client-go [\#80](https://github.com/kubedb/elasticsearch/pull/80) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0](https://github.com/kubedb/elasticsearch/tree/0.1.0) (2017-06-14)
**Fixed bugs:**

- Allow updating to create missing workloads [\#76](https://github.com/kubedb/elasticsearch/pull/76) ([aerokite](https://github.com/aerokite))

**Merged pull requests:**

- Change api version to v1alpha1 [\#79](https://github.com/kubedb/elasticsearch/pull/79) ([tamalsaha](https://github.com/tamalsaha))
- Pass cronController as parameter [\#78](https://github.com/kubedb/elasticsearch/pull/78) ([aerokite](https://github.com/aerokite))
- Use built-in exporter [\#77](https://github.com/kubedb/elasticsearch/pull/77) ([tamalsaha](https://github.com/tamalsaha))
- Add analytics event for operator [\#75](https://github.com/kubedb/elasticsearch/pull/75) ([aerokite](https://github.com/aerokite))
- Add analytics [\#74](https://github.com/kubedb/elasticsearch/pull/74) ([aerokite](https://github.com/aerokite))
- Revendor client-go [\#73](https://github.com/kubedb/elasticsearch/pull/73) ([tamalsaha](https://github.com/tamalsaha))
- Add Run\(\) method to just run controller. [\#72](https://github.com/kubedb/elasticsearch/pull/72) ([tamalsaha](https://github.com/tamalsaha))
- Add HTTP server to expose metrics [\#71](https://github.com/kubedb/elasticsearch/pull/71) ([tamalsaha](https://github.com/tamalsaha))
- Prometheus support [\#70](https://github.com/kubedb/elasticsearch/pull/70) ([saumanbiswas](https://github.com/saumanbiswas))
- Use kubedb docker hub account [\#69](https://github.com/kubedb/elasticsearch/pull/69) ([tamalsaha](https://github.com/tamalsaha))
- Use kubedb instead of k8sdb [\#68](https://github.com/kubedb/elasticsearch/pull/68) ([tamalsaha](https://github.com/tamalsaha))
- Ungroup imports on fmt [\#64](https://github.com/kubedb/elasticsearch/pull/64) ([tamalsaha](https://github.com/tamalsaha))
- Fix go report card issue [\#63](https://github.com/kubedb/elasticsearch/pull/63) ([tamalsaha](https://github.com/tamalsaha))
- Rename DeletedDatabase to DormantDatabase [\#62](https://github.com/kubedb/elasticsearch/pull/62) ([tamalsaha](https://github.com/tamalsaha))
- Fix update operation [\#60](https://github.com/kubedb/elasticsearch/pull/60) ([aerokite](https://github.com/aerokite))
- Remove prefix from snapshot job [\#59](https://github.com/kubedb/elasticsearch/pull/59) ([aerokite](https://github.com/aerokite))
- Rename DatabaseSnapshot to Snapshot [\#58](https://github.com/kubedb/elasticsearch/pull/58) ([tamalsaha](https://github.com/tamalsaha))
- Modify StatefulSet naming format [\#56](https://github.com/kubedb/elasticsearch/pull/56) ([aerokite](https://github.com/aerokite))
- Get object each time before updating [\#55](https://github.com/kubedb/elasticsearch/pull/55) ([aerokite](https://github.com/aerokite))
- Check docker image version [\#54](https://github.com/kubedb/elasticsearch/pull/54) ([aerokite](https://github.com/aerokite))
- Create headless service for StatefulSet [\#53](https://github.com/kubedb/elasticsearch/pull/53) ([aerokite](https://github.com/aerokite))
- Use data as Volume name [\#52](https://github.com/kubedb/elasticsearch/pull/52) ([aerokite](https://github.com/aerokite))
- Use kind in label instead of type [\#50](https://github.com/kubedb/elasticsearch/pull/50) ([aerokite](https://github.com/aerokite))
- Do not store autogenerated meta information [\#49](https://github.com/kubedb/elasticsearch/pull/49) ([aerokite](https://github.com/aerokite))
- Bubble up error for controller methods [\#47](https://github.com/kubedb/elasticsearch/pull/47) ([aerokite](https://github.com/aerokite))
- Modify e2e test. Do not support recovery by recreating Elastic anymore. [\#46](https://github.com/kubedb/elasticsearch/pull/46) ([aerokite](https://github.com/aerokite))
- Use Kubernetes EventRecorder directly [\#45](https://github.com/kubedb/elasticsearch/pull/45) ([aerokite](https://github.com/aerokite))
- Address status field changes [\#44](https://github.com/kubedb/elasticsearch/pull/44) ([aerokite](https://github.com/aerokite))
- Use canary tag for k8sdb images [\#42](https://github.com/kubedb/elasticsearch/pull/42) ([aerokite](https://github.com/aerokite))
- Install ca-certificates in operator docker image. [\#41](https://github.com/kubedb/elasticsearch/pull/41) ([tamalsaha](https://github.com/tamalsaha))
- Add deployment.yaml [\#40](https://github.com/kubedb/elasticsearch/pull/40) ([aerokite](https://github.com/aerokite))
- Rename "destroy" to "wipeOut" [\#38](https://github.com/kubedb/elasticsearch/pull/38) ([tamalsaha](https://github.com/tamalsaha))
- Store Elastic Spec in DeletedDatabase [\#36](https://github.com/kubedb/elasticsearch/pull/36) ([aerokite](https://github.com/aerokite))
- Update timing fields. [\#35](https://github.com/kubedb/elasticsearch/pull/35) ([tamalsaha](https://github.com/tamalsaha))
- Use k8sdb docker hub account [\#34](https://github.com/kubedb/elasticsearch/pull/34) ([tamalsaha](https://github.com/tamalsaha))
- Implement database initialization [\#32](https://github.com/kubedb/elasticsearch/pull/32) ([aerokite](https://github.com/aerokite))
- Use resource name constant from apimachinery [\#31](https://github.com/kubedb/elasticsearch/pull/31) ([tamalsaha](https://github.com/tamalsaha))
- Use one controller struct [\#30](https://github.com/kubedb/elasticsearch/pull/30) ([tamalsaha](https://github.com/tamalsaha))
- Implement updated interfaces. [\#29](https://github.com/kubedb/elasticsearch/pull/29) ([tamalsaha](https://github.com/tamalsaha))
- Rename controller image to k8s-es [\#28](https://github.com/kubedb/elasticsearch/pull/28) ([tamalsaha](https://github.com/tamalsaha))
- Implement Snapshotter, Deleter with Controller [\#27](https://github.com/kubedb/elasticsearch/pull/27) ([aerokite](https://github.com/aerokite))
- Modify implementation [\#26](https://github.com/kubedb/elasticsearch/pull/26) ([aerokite](https://github.com/aerokite))
- Implement interface [\#25](https://github.com/kubedb/elasticsearch/pull/25) ([aerokite](https://github.com/aerokite))
- Reorganize code [\#24](https://github.com/kubedb/elasticsearch/pull/24) ([aerokite](https://github.com/aerokite))
- Modify snapshot name format [\#23](https://github.com/kubedb/elasticsearch/pull/23) ([aerokite](https://github.com/aerokite))
- Modify controller for backup operation [\#22](https://github.com/kubedb/elasticsearch/pull/22) ([aerokite](https://github.com/aerokite))
- Use osm to pull/push snapshots [\#21](https://github.com/kubedb/elasticsearch/pull/21) ([aerokite](https://github.com/aerokite))
- Move api & client to apimachinery [\#20](https://github.com/kubedb/elasticsearch/pull/20) ([aerokite](https://github.com/aerokite))
- Remove DeleteOptions{} [\#18](https://github.com/kubedb/elasticsearch/pull/18) ([aerokite](https://github.com/aerokite))
- Modify labels and annotations [\#17](https://github.com/kubedb/elasticsearch/pull/17) ([aerokite](https://github.com/aerokite))
- Add controller operation [\#16](https://github.com/kubedb/elasticsearch/pull/16) ([aerokite](https://github.com/aerokite))
- Modify types to match TPR "elastic.k8sdb.com" [\#14](https://github.com/kubedb/elasticsearch/pull/14) ([aerokite](https://github.com/aerokite))
- Change Kind "elasticsearch" to "elastic" [\#13](https://github.com/kubedb/elasticsearch/pull/13) ([aerokite](https://github.com/aerokite))
- Move elasticsearch\_discovery & docker files [\#6](https://github.com/kubedb/elasticsearch/pull/6) ([aerokite](https://github.com/aerokite))
- Modify skeleton to elasticsearch [\#4](https://github.com/kubedb/elasticsearch/pull/4) ([aerokite](https://github.com/aerokite))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*