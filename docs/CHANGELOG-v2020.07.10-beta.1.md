---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2020.07.10-beta.1
    name: Changelog-v2020.07.10-beta.1
    parent: welcome
    weight: 20200710
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2020.07.10-beta.1/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2020.07.10-beta.1/
---

# KubeDB v2020.07.10-beta.1 (2020-07-10)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.14.0-beta.1](https://github.com/kubedb/apimachinery/releases/tag/v0.14.0-beta.1)

- [157a8724](https://github.com/kubedb/apimachinery/commit/157a8724) Update for release Stash@v2020.07.09-beta.0 (#541)
- [0e86bdbd](https://github.com/kubedb/apimachinery/commit/0e86bdbd) Update for release Stash@v2020.07.08-beta.0 (#540)
- [f4a22d0c](https://github.com/kubedb/apimachinery/commit/f4a22d0c) Update License notice (#539)
- [3c598500](https://github.com/kubedb/apimachinery/commit/3c598500) Use Allowlist and Denylist in MySQLVersion (#537)
- [3c58c062](https://github.com/kubedb/apimachinery/commit/3c58c062) Update to Kubernetes v1.18.3 (#536)
- [e1f3d603](https://github.com/kubedb/apimachinery/commit/e1f3d603) Update update-release-tracker.sh
- [0cf4a01f](https://github.com/kubedb/apimachinery/commit/0cf4a01f) Update update-release-tracker.sh
- [bfbd1f8d](https://github.com/kubedb/apimachinery/commit/bfbd1f8d) Add script to update release tracker on pr merge (#533)
- [b817d87c](https://github.com/kubedb/apimachinery/commit/b817d87c) Update .kodiak.toml
- [772e8d2f](https://github.com/kubedb/apimachinery/commit/772e8d2f) Add Ops Request const (#529)
- [453d67ca](https://github.com/kubedb/apimachinery/commit/453d67ca) Add constants for mutator & validator group names (#532)
- [69f997b5](https://github.com/kubedb/apimachinery/commit/69f997b5) Unwrap top level api folder (#531)
- [a8ccec51](https://github.com/kubedb/apimachinery/commit/a8ccec51) Make RedisOpsRequest Namespaced (#530)
- [8a076bfb](https://github.com/kubedb/apimachinery/commit/8a076bfb) Update .kodiak.toml
- [6a8e51b9](https://github.com/kubedb/apimachinery/commit/6a8e51b9) Update to Kubernetes v1.18.3 (#527)
- [2ef41962](https://github.com/kubedb/apimachinery/commit/2ef41962) Create .kodiak.toml
- [8e596d4e](https://github.com/kubedb/apimachinery/commit/8e596d4e) Update to Kubernetes v1.18.3
- [31f72200](https://github.com/kubedb/apimachinery/commit/31f72200) Update comments
- [27bc9265](https://github.com/kubedb/apimachinery/commit/27bc9265) Use CRD v1 for Kubernetes >= 1.16 (#525)
- [d1be7d1d](https://github.com/kubedb/apimachinery/commit/d1be7d1d) Remove defaults from CRD v1beta1
- [5c73d507](https://github.com/kubedb/apimachinery/commit/5c73d507) Use crd.Interface in Controller (#524)
- [27763544](https://github.com/kubedb/apimachinery/commit/27763544) Generate both v1beta1 and v1 CRD YAML (#523)
- [5a0f0a93](https://github.com/kubedb/apimachinery/commit/5a0f0a93) Update to Kubernetes v1.18.3 (#520)
- [25008c1a](https://github.com/kubedb/apimachinery/commit/25008c1a) Change MySQL `[]ContainerResources` to `core.ResourceRequirements` (#522)
- [abc99620](https://github.com/kubedb/apimachinery/commit/abc99620) Merge pull request #521 from kubedb/mongo-vertical
- [f38a109c](https://github.com/kubedb/apimachinery/commit/f38a109c) Change `[]ContainerResources` to `core.ResourceRequirements`



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.14.0-beta.1](https://github.com/kubedb/cli/releases/tag/v0.14.0-beta.1)

- [80e77588](https://github.com/kubedb/cli/commit/80e77588) Prepare for release v0.14.0-beta.1 (#468)
- [6925c726](https://github.com/kubedb/cli/commit/6925c726) Update for release Stash@v2020.07.09-beta.0 (#466)
- [6036e14f](https://github.com/kubedb/cli/commit/6036e14f) Update for release Stash@v2020.07.08-beta.0 (#465)
- [03de8e3f](https://github.com/kubedb/cli/commit/03de8e3f) Disable autogen tags in docs (#464)
- [3bcfa7ef](https://github.com/kubedb/cli/commit/3bcfa7ef) Update License (#463)
- [0aa91f93](https://github.com/kubedb/cli/commit/0aa91f93) Update to Kubernetes v1.18.3 (#462)
- [023555ef](https://github.com/kubedb/cli/commit/023555ef) Add workflow to update docs (#461)
- [abd9d054](https://github.com/kubedb/cli/commit/abd9d054) Update update-release-tracker.sh
- [0a9527d4](https://github.com/kubedb/cli/commit/0a9527d4) Update update-release-tracker.sh
- [69c644a2](https://github.com/kubedb/cli/commit/69c644a2) Add script to update release tracker on pr merge (#460)
- [595679ba](https://github.com/kubedb/cli/commit/595679ba) Make release non-draft
- [880d3492](https://github.com/kubedb/cli/commit/880d3492) Update .kodiak.toml
- [a7607798](https://github.com/kubedb/cli/commit/a7607798) Update to Kubernetes v1.18.3 (#459)
- [3197b4b7](https://github.com/kubedb/cli/commit/3197b4b7) Update to Kubernetes v1.18.3
- [8ed52c84](https://github.com/kubedb/cli/commit/8ed52c84) Create .kodiak.toml
- [cfda68d4](https://github.com/kubedb/cli/commit/cfda68d4) Update to Kubernetes v1.18.3 (#458)
- [7395c039](https://github.com/kubedb/cli/commit/7395c039) Update dependencies
- [542e6709](https://github.com/kubedb/cli/commit/542e6709) Update crazy-max/ghaction-docker-buildx flag
- [972d8119](https://github.com/kubedb/cli/commit/972d8119) Revendor kubedb.dev/apimachinery@master
- [540e5a7d](https://github.com/kubedb/cli/commit/540e5a7d) Cleanup cli commands (#454)
- [98649b0a](https://github.com/kubedb/cli/commit/98649b0a) Trigger the workflow on push or pull request
- [a0dbdab5](https://github.com/kubedb/cli/commit/a0dbdab5) Update readme (#457)
- [a52927ed](https://github.com/kubedb/cli/commit/a52927ed) Create draft GitHub release when tagged (#456)
- [42838aec](https://github.com/kubedb/cli/commit/42838aec) Convert kubedb cli into a `kubectl dba` plgin (#455)
- [aec37df2](https://github.com/kubedb/cli/commit/aec37df2) Revendor dependencies
- [2c120d1a](https://github.com/kubedb/cli/commit/2c120d1a) Update client-go to kubernetes-1.16.3 (#453)
- [ce221024](https://github.com/kubedb/cli/commit/ce221024) Add add-license make target
- [84a6a1e8](https://github.com/kubedb/cli/commit/84a6a1e8) Add license header to files (#452)
- [1ced65ea](https://github.com/kubedb/cli/commit/1ced65ea) Split imports into 3 parts (#451)
- [8e533f69](https://github.com/kubedb/cli/commit/8e533f69) Add release workflow script (#450)
- [0735ce0c](https://github.com/kubedb/cli/commit/0735ce0c) Enable GitHub actions
- [8522ec74](https://github.com/kubedb/cli/commit/8522ec74) Update changelog



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.14.0-beta.1](https://github.com/kubedb/elasticsearch/releases/tag/v0.14.0-beta.1)

- [9aae4782](https://github.com/kubedb/elasticsearch/commit/9aae4782) Prepare for release v0.14.0-beta.1 (#319)
- [312e5682](https://github.com/kubedb/elasticsearch/commit/312e5682) Update for release Stash@v2020.07.09-beta.0 (#317)
- [681f3e87](https://github.com/kubedb/elasticsearch/commit/681f3e87) Include Makefile.env
- [e460af51](https://github.com/kubedb/elasticsearch/commit/e460af51) Allow customizing chart registry (#316)
- [64e15a33](https://github.com/kubedb/elasticsearch/commit/64e15a33) Update for release Stash@v2020.07.08-beta.0 (#315)
- [1f2ef7a6](https://github.com/kubedb/elasticsearch/commit/1f2ef7a6) Update License (#314)
- [16ce6c90](https://github.com/kubedb/elasticsearch/commit/16ce6c90) Update to Kubernetes v1.18.3 (#313)
- [3357faa3](https://github.com/kubedb/elasticsearch/commit/3357faa3) Update ci.yml
- [cb44a1eb](https://github.com/kubedb/elasticsearch/commit/cb44a1eb) Load stash version from .env file for make (#312)
- [cf212019](https://github.com/kubedb/elasticsearch/commit/cf212019) Update update-release-tracker.sh
- [5127428e](https://github.com/kubedb/elasticsearch/commit/5127428e) Update update-release-tracker.sh
- [7f790940](https://github.com/kubedb/elasticsearch/commit/7f790940) Add script to update release tracker on pr merge (#311)
- [340b6112](https://github.com/kubedb/elasticsearch/commit/340b6112) Update .kodiak.toml
- [e01c4eec](https://github.com/kubedb/elasticsearch/commit/e01c4eec) Various fixes (#310)
- [11517f71](https://github.com/kubedb/elasticsearch/commit/11517f71) Update to Kubernetes v1.18.3 (#309)
- [53d7b117](https://github.com/kubedb/elasticsearch/commit/53d7b117) Update to Kubernetes v1.18.3
- [7eacc7dd](https://github.com/kubedb/elasticsearch/commit/7eacc7dd) Create .kodiak.toml
- [b91b23d9](https://github.com/kubedb/elasticsearch/commit/b91b23d9) Use CRD v1 for Kubernetes >= 1.16 (#308)
- [08c1d2a8](https://github.com/kubedb/elasticsearch/commit/08c1d2a8) Update to Kubernetes v1.18.3 (#307)
- [32cdb8a4](https://github.com/kubedb/elasticsearch/commit/32cdb8a4) Fix e2e tests (#306)
- [0bca1a04](https://github.com/kubedb/elasticsearch/commit/0bca1a04) Merge pull request #302 from kubedb/multi-region
- [bf0c26ee](https://github.com/kubedb/elasticsearch/commit/bf0c26ee) Revendor kubedb.dev/apimachinery@v0.14.0-beta.0
- [7c00c63c](https://github.com/kubedb/elasticsearch/commit/7c00c63c) Add support for multi-regional cluster
- [363322df](https://github.com/kubedb/elasticsearch/commit/363322df) Update stash install commands
- [a0138a36](https://github.com/kubedb/elasticsearch/commit/a0138a36) Update crazy-max/ghaction-docker-buildx flag
- [3076eb46](https://github.com/kubedb/elasticsearch/commit/3076eb46) Use updated operator labels in e2e tests (#304)
- [d537b91b](https://github.com/kubedb/elasticsearch/commit/d537b91b) Pass annotations from CRD to AppBinding (#305)
- [48f9399c](https://github.com/kubedb/elasticsearch/commit/48f9399c) Trigger the workflow on push or pull request
- [7b8d56cb](https://github.com/kubedb/elasticsearch/commit/7b8d56cb) Update CHANGELOG.md
- [939f6882](https://github.com/kubedb/elasticsearch/commit/939f6882) Update labelSelector for statefulsets (#300)
- [ed1c0553](https://github.com/kubedb/elasticsearch/commit/ed1c0553) Make master service headless & add rest-port to all db nodes (#299)
- [b7e7c8d7](https://github.com/kubedb/elasticsearch/commit/b7e7c8d7) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#301)
- [e51555d5](https://github.com/kubedb/elasticsearch/commit/e51555d5) Introduce spec.halted and removed dormant and snapshot crd (#296)
- [8255276f](https://github.com/kubedb/elasticsearch/commit/8255276f) Add spec.selector fields to the governing service (#297)
- [13bc760f](https://github.com/kubedb/elasticsearch/commit/13bc760f) Use stash@v0.9.0-rc.4 release (#298)
- [6a21fb86](https://github.com/kubedb/elasticsearch/commit/6a21fb86) Add `Pause` feature (#295)
- [1b25070c](https://github.com/kubedb/elasticsearch/commit/1b25070c) Refactor CI pipeline to build once (#294)
- [ace3d779](https://github.com/kubedb/elasticsearch/commit/ace3d779) Fix e2e tests on GitHub actions (#292)
- [7a7eb8d1](https://github.com/kubedb/elasticsearch/commit/7a7eb8d1) fix bug (#293)
- [0641649e](https://github.com/kubedb/elasticsearch/commit/0641649e) Use Go 1.13 in CI (#291)
- [97790e1e](https://github.com/kubedb/elasticsearch/commit/97790e1e) Take out elasticsearch docker images and Matrix test (#289)
- [3a20c1db](https://github.com/kubedb/elasticsearch/commit/3a20c1db) Fix default make command
- [ece073a2](https://github.com/kubedb/elasticsearch/commit/ece073a2) Update catalog values for make install command
- [8df4697b](https://github.com/kubedb/elasticsearch/commit/8df4697b) Use charts to install operator (#290)
- [5cbde391](https://github.com/kubedb/elasticsearch/commit/5cbde391) Add add-license make target
- [b7012bc5](https://github.com/kubedb/elasticsearch/commit/b7012bc5) Skip libbuild folder from checking license
- [d56db3a0](https://github.com/kubedb/elasticsearch/commit/d56db3a0) Add license header to files (#288)
- [1d0c368a](https://github.com/kubedb/elasticsearch/commit/1d0c368a) Enable make ci (#287)
- [2e835dff](https://github.com/kubedb/elasticsearch/commit/2e835dff) Remove EnableStatusSubresource (#286)
- [bcd0ebd9](https://github.com/kubedb/elasticsearch/commit/bcd0ebd9) Fix E2E tests in github action (#285)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.14.0-beta.1](https://github.com/kubedb/installer/releases/tag/v0.14.0-beta.1)

- [a081a36](https://github.com/kubedb/installer/commit/a081a36) Prepare for release v0.14.0-beta.1 (#107)
- [9c3fd4a](https://github.com/kubedb/installer/commit/9c3fd4a) Make chart registry configurable (#106)
- [a3da9a1](https://github.com/kubedb/installer/commit/a3da9a1) Publish to testing dir for alpha/beta releases
- [33685ee](https://github.com/kubedb/installer/commit/33685ee) Update License (#105)
- [f06fa20](https://github.com/kubedb/installer/commit/f06fa20) Update MySQL version catalog (#104)
- [674d129](https://github.com/kubedb/installer/commit/674d129) Update to Kubernetes v1.18.3 (#101)
- [fc16306](https://github.com/kubedb/installer/commit/fc16306) Update ci.yml
- [f65dd16](https://github.com/kubedb/installer/commit/f65dd16) Tag chart and app version as string for yq
- [ac21db4](https://github.com/kubedb/installer/commit/ac21db4) Update links (#100)
- [4a71c15](https://github.com/kubedb/installer/commit/4a71c15) Update update-release-tracker.sh
- [e7f14e9](https://github.com/kubedb/installer/commit/e7f14e9) Update update-release-tracker.sh
- [b26d3b8](https://github.com/kubedb/installer/commit/b26d3b8) Update release.yml
- [4f4985d](https://github.com/kubedb/installer/commit/4f4985d) Add script to update release tracker on pr merge (#98)
- [94baab8](https://github.com/kubedb/installer/commit/94baab8) Update ci.yml
- [2ffe241](https://github.com/kubedb/installer/commit/2ffe241) Rename TEST_NAMESPACE -> KUBE_NAMESPACE
- [34ba017](https://github.com/kubedb/installer/commit/34ba017) Change Enterprise operator image name to kubedb-enterprise (#97)
- [bc83b11](https://github.com/kubedb/installer/commit/bc83b11) Add commands to update chart (#96)
- [a0ddc4b](https://github.com/kubedb/installer/commit/a0ddc4b) Bring back postgres 9.6 (#95)
- [59c1cee](https://github.com/kubedb/installer/commit/59c1cee) Fix chart release process (#94)
- [40072c6](https://github.com/kubedb/installer/commit/40072c6) Deprecate non-patched versions (#93)
- [16f09ed](https://github.com/kubedb/installer/commit/16f09ed) Update .kodiak.toml
- [bb902e3](https://github.com/kubedb/installer/commit/bb902e3) Release kubedb-enterprise chart to stable charts
- [7c94dfc](https://github.com/kubedb/installer/commit/7c94dfc) Remove default deprecated: false fields (#92)
- [07b162d](https://github.com/kubedb/installer/commit/07b162d) Update chart versions (#91)
- [dd156da](https://github.com/kubedb/installer/commit/dd156da) Add rbac for configmaps
- [a175cc9](https://github.com/kubedb/installer/commit/a175cc9) Revise the validator & mutator webhook names (#90)
- [777b636](https://github.com/kubedb/installer/commit/777b636) Add kubedb-enterprise chart (#89)
- [6d4f4d8](https://github.com/kubedb/installer/commit/6d4f4d8) Update to Kubernetes v1.18.3 (#84)
- [8065729](https://github.com/kubedb/installer/commit/8065729) Update to Kubernetes v1.18.3
- [87052ae](https://github.com/kubedb/installer/commit/87052ae) Create .kodiak.toml
- [8c8c122](https://github.com/kubedb/installer/commit/8c8c122) Add RBAC permission for generic garbage collector (#82)
- [f391304](https://github.com/kubedb/installer/commit/f391304) Permit configmap list/watch for delegated authentication (#81)
- [96dbad6](https://github.com/kubedb/installer/commit/96dbad6) Use updated image registry values field
- [a770d06](https://github.com/kubedb/installer/commit/a770d06) Generate both v1beta1 and v1 CRD YAML (#80)
- [ee01bf6](https://github.com/kubedb/installer/commit/ee01bf6) Update to Kubernetes v1.18.3 (#79)
- [7e6edc3](https://github.com/kubedb/installer/commit/7e6edc3) Update chart docs
- [71e999d](https://github.com/kubedb/installer/commit/71e999d) Remove combined redis catalog template
- [19aa0a1](https://github.com/kubedb/installer/commit/19aa0a1) Merge pull request #77 from kubedb/opsvalidator
- [103ca84](https://github.com/kubedb/installer/commit/103ca84) Use enterprise port values
- [5d538b8](https://github.com/kubedb/installer/commit/5d538b8) Add ops request validator
- [ce37683](https://github.com/kubedb/installer/commit/ce37683) Update Enterprise operator tag (#78)
- [9a08d70](https://github.com/kubedb/installer/commit/9a08d70) Merge pull request #76 from kubedb/mysqlnewversion
- [82a2d67](https://github.com/kubedb/installer/commit/82a2d67) remove unnecessary code and rename standAlone to standalone
- [f3f6d05](https://github.com/kubedb/installer/commit/f3f6d05) Add extra wrap for depricated version
- [7206194](https://github.com/kubedb/installer/commit/7206194) Add mysql new version
- [f4e79c8](https://github.com/kubedb/installer/commit/f4e79c8) Rename api group to ops.kubedb.com (#75)
- [ee49da5](https://github.com/kubedb/installer/commit/ee49da5) Add skipDeprecated to catalog chart (#74)
- [dd6d4f9](https://github.com/kubedb/installer/commit/dd6d4f9) Split db catalog into separate files per version (#73)
- [4ab187b](https://github.com/kubedb/installer/commit/4ab187b) Merge pull request #71 from kubedb/fix-ci
- [bdbc6b5](https://github.com/kubedb/installer/commit/bdbc6b5) Remove PSP for Snapshot
- [b51576b](https://github.com/kubedb/installer/commit/b51576b) Use recommended kubernetes app labels
- [6f5a51c](https://github.com/kubedb/installer/commit/6f5a51c) Merge pull request #72 from pohly/memached-1.5.22
- [a89d9bf](https://github.com/kubedb/installer/commit/a89d9bf) memcached: add 1.5.22
- [a6b63d6](https://github.com/kubedb/installer/commit/a6b63d6) Trigger the workflow on push or pull request
- [600eb93](https://github.com/kubedb/installer/commit/600eb93) Update chart readme
- [df4bcb2](https://github.com/kubedb/installer/commit/df4bcb2) Auto generate chart readme file
- [00ca986](https://github.com/kubedb/installer/commit/00ca986) Use GCR_SERVICE_ACCOUNT_JSON_KEY env in CI
- [c0cdfe0](https://github.com/kubedb/installer/commit/c0cdfe0) Configure Docker credential helper
- [06ed3df](https://github.com/kubedb/installer/commit/06ed3df) Use gcr.io/appscode to host Enterprise operator image
- [9d0fbc9](https://github.com/kubedb/installer/commit/9d0fbc9) Update release.yml
- [e066043](https://github.com/kubedb/installer/commit/e066043) prometheus.io/coreos-operator -> prometheus.io/coreos-operator (#66)
- [91f37ec](https://github.com/kubedb/installer/commit/91f37ec) Use image.registry in catalog chart (#65)
- [a1ad35c](https://github.com/kubedb/installer/commit/a1ad35c) Move apireg annotation to operator pod (#64)
- [b02b054](https://github.com/kubedb/installer/commit/b02b054) Add fuzz tests for CRDs (#63)
- [12c1d4f](https://github.com/kubedb/installer/commit/12c1d4f) Various fixes (#62)
- [9b572fa](https://github.com/kubedb/installer/commit/9b572fa) Use kubectl v1.17.0 (#61)
- [2825f18](https://github.com/kubedb/installer/commit/2825f18) Fix helm install --wait flag (#57)
- [3e205ae](https://github.com/kubedb/installer/commit/3e205ae) Fix tolerations indentation for deployment (#58)
- [bed096d](https://github.com/kubedb/installer/commit/bed096d) Add cluster-role for dba.kubedb.com (#54)
- [a684c02](https://github.com/kubedb/installer/commit/a684c02) Update user roles for KubeDB crds (#60)
- [9e4f924](https://github.com/kubedb/installer/commit/9e4f924) Add release script to upload charts (#55)
- [1a9ba37](https://github.com/kubedb/installer/commit/1a9ba37) Updated Mongodb Init images (#51)
- [1ab4bed](https://github.com/kubedb/installer/commit/1ab4bed) Run checks once in CI pipeline (#53)
- [5265527](https://github.com/kubedb/installer/commit/5265527) Properly mark optional fields (#52)
- [6153622](https://github.com/kubedb/installer/commit/6153622) Add replicationModeDetector image field into MySQLVersion CRD (#50)
- [d546169](https://github.com/kubedb/installer/commit/d546169) Add Enterprise operator sidecar (#49)
- [cbc7f03](https://github.com/kubedb/installer/commit/cbc7f03) Add deletocollection verbs to kubedb roles (#44)
- [c598e90](https://github.com/kubedb/installer/commit/c598e90) Allow specifying rather than generating certs (#48)
- [b39d710](https://github.com/kubedb/installer/commit/b39d710) RBAC for cert manger, issuer watcher, and secret watcher (#43)
- [0276c34](https://github.com/kubedb/installer/commit/0276c34) Add missing permissions for PgBouncer operator (#47)
- [b85efed](https://github.com/kubedb/installer/commit/b85efed) Update Installer for ProxySQL and PerconaXtraDB (#46)
- [5c8212a](https://github.com/kubedb/installer/commit/5c8212a) Don't install PSP policy when catalog is disabled. (#45)
- [b8ebcdb](https://github.com/kubedb/installer/commit/b8ebcdb) Bring back support for k8s 1.11 (#42)
- [b9453f2](https://github.com/kubedb/installer/commit/b9453f2) Change minimum k8s req to 1.12 and use helm 3 in chart readme (#41)
- [29b4a96](https://github.com/kubedb/installer/commit/29b4a96) Add catalog for percona standalone (#40)
- [1ff9a1f](https://github.com/kubedb/installer/commit/1ff9a1f) Avoid creating apiservices when webhooks are disabled (#39)
- [b96eeba](https://github.com/kubedb/installer/commit/b96eeba) Update kubedb-catalog values
- [c6bc91a](https://github.com/kubedb/installer/commit/c6bc91a) Conditionally create validating and mutating webhooks. (#38)
- [e390c04](https://github.com/kubedb/installer/commit/e390c04) Delete script based installer (#36)
- [ab0f799](https://github.com/kubedb/installer/commit/ab0f799) Update installer for ProxySQL (#17)
- [5762cbf](https://github.com/kubedb/installer/commit/5762cbf) Update installer for PerconaXtraDB (#14)
- [6b5565a](https://github.com/kubedb/installer/commit/6b5565a) Pass imagePullSecrets as an array to service accounts (#37)
- [3a552a1](https://github.com/kubedb/installer/commit/3a552a1) Use helmpack/chart-testing:v3.0.0-beta.1 (#35)
- [13fc00b](https://github.com/kubedb/installer/commit/13fc00b) Fix RBAC permissions for Stash restoresessions (#34)
- [0023d58](https://github.com/kubedb/installer/commit/0023d58) Mark optional fields in installer CRD
- [b24b05e](https://github.com/kubedb/installer/commit/b24b05e) Add installer api CRD (#31)
- [51f80ea](https://github.com/kubedb/installer/commit/51f80ea) Always create rbac resources (#32)
- [f36f6c8](https://github.com/kubedb/installer/commit/f36f6c8) Use kind v0.6.1 (#30)
- [b13266e](https://github.com/kubedb/installer/commit/b13266e) Properly handle empty image pull secret name in installer (#29)
- [0243c9e](https://github.com/kubedb/installer/commit/0243c9e) Test installers (#27)
- [5aaba63](https://github.com/kubedb/installer/commit/5aaba63) Fix typo (#28)
- [dd2595d](https://github.com/kubedb/installer/commit/dd2595d) Use separate docker registry for operator and catalog images (#26)
- [316f340](https://github.com/kubedb/installer/commit/316f340) Use pgbouncer_exporter:v0.1.1
- [29843a0](https://github.com/kubedb/installer/commit/29843a0) Support for pgbouncers (#11)
- [2f2f902](https://github.com/kubedb/installer/commit/2f2f902) Ensure operator service points to its own pod. (#25)
- [d187265](https://github.com/kubedb/installer/commit/d187265) Update postgres versions (#24)
- [167fe46](https://github.com/kubedb/installer/commit/167fe46) Remove --enable-status-subresource flag (#23)
- [025afcb](https://github.com/kubedb/installer/commit/025afcb) ESVerdion 7.3.2 and 7.3 added (#21)
- [adb433f](https://github.com/kubedb/installer/commit/adb433f) Support for xpack in es6.8 and es7.2 (#20)
- [22634fa](https://github.com/kubedb/installer/commit/22634fa) Add crd for elasticsearch 7.2.0 (#9)
- [b06b2ea](https://github.com/kubedb/installer/commit/b06b2ea) Add namespace to cleaner Job (#18)
- [ac173e6](https://github.com/kubedb/installer/commit/ac173e6) Download onessl version v0.13.1 for Kubernetes 1.16 fix (#19)
- [3375df9](https://github.com/kubedb/installer/commit/3375df9) Use percona mongodb exporter from 0.13.0 (#16)
- [fdc6105](https://github.com/kubedb/installer/commit/fdc6105) Add support for Elasticsearch 6.8.0 (#7)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.7.0-beta.1](https://github.com/kubedb/memcached/releases/tag/v0.7.0-beta.1)

- [3f7c1b90](https://github.com/kubedb/memcached/commit/3f7c1b90) Prepare for release v0.7.0-beta.1 (#160)
- [1278cd57](https://github.com/kubedb/memcached/commit/1278cd57) include Makefile.env (#158)
- [676222b7](https://github.com/kubedb/memcached/commit/676222b7) Update License (#157)
- [216fdcd4](https://github.com/kubedb/memcached/commit/216fdcd4) Update to Kubernetes v1.18.3 (#156)
- [dc59abf4](https://github.com/kubedb/memcached/commit/dc59abf4) Update ci.yml
- [071589c5](https://github.com/kubedb/memcached/commit/071589c5) Update update-release-tracker.sh
- [79bc96d8](https://github.com/kubedb/memcached/commit/79bc96d8) Update update-release-tracker.sh
- [31f5fca6](https://github.com/kubedb/memcached/commit/31f5fca6) Add script to update release tracker on pr merge (#155)
- [05d1d6ab](https://github.com/kubedb/memcached/commit/05d1d6ab) Update .kodiak.toml
- [522b617f](https://github.com/kubedb/memcached/commit/522b617f) Various fixes (#154)
- [2ed2c3a0](https://github.com/kubedb/memcached/commit/2ed2c3a0) Update to Kubernetes v1.18.3 (#152)
- [10cea9ad](https://github.com/kubedb/memcached/commit/10cea9ad) Update to Kubernetes v1.18.3
- [582177b0](https://github.com/kubedb/memcached/commit/582177b0) Create .kodiak.toml
- [bf1900b6](https://github.com/kubedb/memcached/commit/bf1900b6) Run flaky e2e test (#151)
- [aa09abfc](https://github.com/kubedb/memcached/commit/aa09abfc) Use CRD v1 for Kubernetes >= 1.16 (#150)
- [b2586151](https://github.com/kubedb/memcached/commit/b2586151) Merge pull request #146 from pohly/pmem
- [dbd5b2b0](https://github.com/kubedb/memcached/commit/dbd5b2b0) Fix build
- [d0722c34](https://github.com/kubedb/memcached/commit/d0722c34) WIP: implement PMEM support
- [f16b1198](https://github.com/kubedb/memcached/commit/f16b1198) Makefile: adapt to recent installer repo changes
- [32f71c56](https://github.com/kubedb/memcached/commit/32f71c56) Makefile: support e2e testing with arbitrary KUBECONFIG file
- [6ed07efc](https://github.com/kubedb/memcached/commit/6ed07efc) Update to Kubernetes v1.18.3 (#149)
- [ce702669](https://github.com/kubedb/memcached/commit/ce702669) Fix e2e tests (#148)
- [18917f8d](https://github.com/kubedb/memcached/commit/18917f8d) Revendor kubedb.dev/apimachinery@master (#147)
- [e51d327c](https://github.com/kubedb/memcached/commit/e51d327c) Update crazy-max/ghaction-docker-buildx flag
- [1202c059](https://github.com/kubedb/memcached/commit/1202c059) Use updated operator labels in e2e tests (#144)
- [e02d42a4](https://github.com/kubedb/memcached/commit/e02d42a4) Pass annotations from CRD to AppBinding (#145)
- [2c91d63b](https://github.com/kubedb/memcached/commit/2c91d63b) Trigger the workflow on push or pull request
- [67c83a9a](https://github.com/kubedb/memcached/commit/67c83a9a) Update CHANGELOG.md
- [85e3cf54](https://github.com/kubedb/memcached/commit/85e3cf54) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#143)
- [e61dd2e6](https://github.com/kubedb/memcached/commit/e61dd2e6) Update error msg to reject halt when termination policy is 'DoNotTerminate'
- [bc079b7b](https://github.com/kubedb/memcached/commit/bc079b7b) Introduce spec.halted and removed dormant crd (#142)
- [f31610c3](https://github.com/kubedb/memcached/commit/f31610c3) Refactor CI pipeline to run build once (#141)
- [f5eec5e4](https://github.com/kubedb/memcached/commit/f5eec5e4) Update kubernetes client-go to 1.16.3 (#140)
- [f645174a](https://github.com/kubedb/memcached/commit/f645174a) Update catalog values for make install command
- [2a297c89](https://github.com/kubedb/memcached/commit/2a297c89) Use charts to install operator (#139)
- [83e2ba17](https://github.com/kubedb/memcached/commit/83e2ba17) Moved out docker files and added matrix github actions ci/cd (#138)
- [97e3a5bd](https://github.com/kubedb/memcached/commit/97e3a5bd) Add add-license make target
- [7b79fbfe](https://github.com/kubedb/memcached/commit/7b79fbfe) Add license header to files (#137)
- [2afa406f](https://github.com/kubedb/memcached/commit/2afa406f) Enable make ci (#136)
- [bab32534](https://github.com/kubedb/memcached/commit/bab32534) Remove EnableStatusSubresource (#135)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.7.0-beta.1](https://github.com/kubedb/mongodb/releases/tag/v0.7.0-beta.1)

- [b82a8fa7](https://github.com/kubedb/mongodb/commit/b82a8fa7) Prepare for release v0.7.0-beta.1 (#211)
- [a63d53ae](https://github.com/kubedb/mongodb/commit/a63d53ae) Update for release Stash@v2020.07.09-beta.0 (#209)
- [4e33e978](https://github.com/kubedb/mongodb/commit/4e33e978) include Makefile.env
- [1aa81a18](https://github.com/kubedb/mongodb/commit/1aa81a18) Allow customizing chart registry (#208)
- [05355e75](https://github.com/kubedb/mongodb/commit/05355e75) Update for release Stash@v2020.07.08-beta.0 (#207)
- [4f6be7b4](https://github.com/kubedb/mongodb/commit/4f6be7b4) Update License (#206)
- [cc54f7d3](https://github.com/kubedb/mongodb/commit/cc54f7d3) Update to Kubernetes v1.18.3 (#204)
- [d1a51b8e](https://github.com/kubedb/mongodb/commit/d1a51b8e) Update ci.yml
- [3a993329](https://github.com/kubedb/mongodb/commit/3a993329) Load stash version from .env file for make (#203)
- [7180a98c](https://github.com/kubedb/mongodb/commit/7180a98c) Update update-release-tracker.sh
- [745085fd](https://github.com/kubedb/mongodb/commit/745085fd) Update update-release-tracker.sh
- [07d83ac0](https://github.com/kubedb/mongodb/commit/07d83ac0) Add script to update release tracker on pr merge (#202)
- [bbe205bb](https://github.com/kubedb/mongodb/commit/bbe205bb) Update .kodiak.toml
- [998e656e](https://github.com/kubedb/mongodb/commit/998e656e) Various fixes (#201)
- [ca03db09](https://github.com/kubedb/mongodb/commit/ca03db09) Update to Kubernetes v1.18.3 (#200)
- [975fc700](https://github.com/kubedb/mongodb/commit/975fc700) Update to Kubernetes v1.18.3
- [52972dcf](https://github.com/kubedb/mongodb/commit/52972dcf) Create .kodiak.toml
- [39168e53](https://github.com/kubedb/mongodb/commit/39168e53) Use CRD v1 for Kubernetes >= 1.16 (#199)
- [d6d87e16](https://github.com/kubedb/mongodb/commit/d6d87e16) Update to Kubernetes v1.18.3 (#198)
- [09cd5809](https://github.com/kubedb/mongodb/commit/09cd5809) Fix e2e tests (#197)
- [f47c4846](https://github.com/kubedb/mongodb/commit/f47c4846) Update stash install commands
- [010d0294](https://github.com/kubedb/mongodb/commit/010d0294) Revendor kubedb.dev/apimachinery@master (#196)
- [31ef2632](https://github.com/kubedb/mongodb/commit/31ef2632) Pass annotations from CRD to AppBinding (#195)
- [9594e92f](https://github.com/kubedb/mongodb/commit/9594e92f) Update crazy-max/ghaction-docker-buildx flag
- [0693d7a0](https://github.com/kubedb/mongodb/commit/0693d7a0) Use updated operator labels in e2e tests (#193)
- [5aaeeb90](https://github.com/kubedb/mongodb/commit/5aaeeb90) Trigger the workflow on push or pull request
- [2af16e3c](https://github.com/kubedb/mongodb/commit/2af16e3c) Update CHANGELOG.md
- [288c5d2f](https://github.com/kubedb/mongodb/commit/288c5d2f) Use SHARD_INDEX constant from apimachinery
- [4482edf3](https://github.com/kubedb/mongodb/commit/4482edf3) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#191)
- [0f20ff3a](https://github.com/kubedb/mongodb/commit/0f20ff3a) Manage SSL certificates using cert-manager (#190)
- [6f0c1aef](https://github.com/kubedb/mongodb/commit/6f0c1aef) Use Minio storage for testing (#188)
- [f8c56bac](https://github.com/kubedb/mongodb/commit/f8c56bac) Support affinity templating in mongodb-shard (#186)
- [71283767](https://github.com/kubedb/mongodb/commit/71283767) Use stash@v0.9.0-rc.4 release (#185)
- [f480de35](https://github.com/kubedb/mongodb/commit/f480de35) Fix `Pause` Logic (#184)
- [263e1bac](https://github.com/kubedb/mongodb/commit/263e1bac) Refactor CI pipeline to build once (#182)
- [e383f271](https://github.com/kubedb/mongodb/commit/e383f271) Add `Pause` Feature (#181)
- [584ecde6](https://github.com/kubedb/mongodb/commit/584ecde6) Delete backupconfig before attempting restoresession. (#180)
- [a78bc2a7](https://github.com/kubedb/mongodb/commit/a78bc2a7) Wipeout if custom databaseSecret has been deleted (#179)
- [e90cd386](https://github.com/kubedb/mongodb/commit/e90cd386) Matrix test and Moved out mongo docker files (#178)
- [c132db8f](https://github.com/kubedb/mongodb/commit/c132db8f) Add add-license makefile target
- [cc545e04](https://github.com/kubedb/mongodb/commit/cc545e04) Update Makefile
- [7a2eab2c](https://github.com/kubedb/mongodb/commit/7a2eab2c) Add license header to files (#177)
- [eecdb2cb](https://github.com/kubedb/mongodb/commit/eecdb2cb) Fix E2E tests in github action (#176)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.7.0-beta.1](https://github.com/kubedb/mysql/releases/tag/v0.7.0-beta.1)

- [19ccc5b8](https://github.com/kubedb/mysql/commit/19ccc5b8) Prepare for release v0.7.0-beta.1 (#201)
- [e61de0e7](https://github.com/kubedb/mysql/commit/e61de0e7) Update for release Stash@v2020.07.09-beta.0 (#199)
- [3269df76](https://github.com/kubedb/mysql/commit/3269df76) Allow customizing chart registry (#198)
- [c487e68e](https://github.com/kubedb/mysql/commit/c487e68e) Update for release Stash@v2020.07.08-beta.0 (#197)
- [4f288ef0](https://github.com/kubedb/mysql/commit/4f288ef0) Update License (#196)
- [858a5e03](https://github.com/kubedb/mysql/commit/858a5e03) Update to Kubernetes v1.18.3 (#195)
- [88dec378](https://github.com/kubedb/mysql/commit/88dec378) Update ci.yml
- [31ef7c2a](https://github.com/kubedb/mysql/commit/31ef7c2a) Load stash version from .env file for make (#194)
- [872954a9](https://github.com/kubedb/mysql/commit/872954a9) Update update-release-tracker.sh
- [771059b9](https://github.com/kubedb/mysql/commit/771059b9) Update update-release-tracker.sh
- [0e625902](https://github.com/kubedb/mysql/commit/0e625902) Add script to update release tracker on pr merge (#193)
- [6a204efd](https://github.com/kubedb/mysql/commit/6a204efd) Update .kodiak.toml
- [de6fc09b](https://github.com/kubedb/mysql/commit/de6fc09b) Various fixes (#192)
- [86eb3313](https://github.com/kubedb/mysql/commit/86eb3313) Update to Kubernetes v1.18.3 (#191)
- [937afcc8](https://github.com/kubedb/mysql/commit/937afcc8) Update to Kubernetes v1.18.3
- [8646a9c8](https://github.com/kubedb/mysql/commit/8646a9c8) Create .kodiak.toml
- [9f3d2e3c](https://github.com/kubedb/mysql/commit/9f3d2e3c) Use helm --wait in make install command
- [3d1e9cf3](https://github.com/kubedb/mysql/commit/3d1e9cf3) Use CRD v1 for Kubernetes >= 1.16 (#188)
- [5df90daa](https://github.com/kubedb/mysql/commit/5df90daa) Merge pull request #187 from kubedb/k-1.18.3
- [179207de](https://github.com/kubedb/mysql/commit/179207de) Pass context
- [76c3fc86](https://github.com/kubedb/mysql/commit/76c3fc86) Update to Kubernetes v1.18.3
- [da9ad307](https://github.com/kubedb/mysql/commit/da9ad307) Fix e2e tests (#186)
- [d7f2c63d](https://github.com/kubedb/mysql/commit/d7f2c63d) Update stash install commands
- [cfee601b](https://github.com/kubedb/mysql/commit/cfee601b) Revendor kubedb.dev/apimachinery@master (#185)
- [741fada4](https://github.com/kubedb/mysql/commit/741fada4) Update crazy-max/ghaction-docker-buildx flag
- [27291b98](https://github.com/kubedb/mysql/commit/27291b98) Use updated operator labels in e2e tests (#183)
- [16b00f9d](https://github.com/kubedb/mysql/commit/16b00f9d) Pass annotations from CRD to AppBinding (#184)
- [b70e0620](https://github.com/kubedb/mysql/commit/b70e0620) Trigger the workflow on push or pull request
- [6ea308d8](https://github.com/kubedb/mysql/commit/6ea308d8) Update CHANGELOG.md
- [188c3a91](https://github.com/kubedb/mysql/commit/188c3a91) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#181)
- [f4a67e95](https://github.com/kubedb/mysql/commit/f4a67e95) Introduce spec.halted and removed dormant and snapshot crd (#178)
- [8774a90c](https://github.com/kubedb/mysql/commit/8774a90c) Use stash@v0.9.0-rc.4 release (#179)
- [209653e6](https://github.com/kubedb/mysql/commit/209653e6) Use apache thrift v0.13.0
- [e89fbe40](https://github.com/kubedb/mysql/commit/e89fbe40) Update github.com/apache/thrift v0.12.0 (#176)
- [c0d035c9](https://github.com/kubedb/mysql/commit/c0d035c9) Add Pause Feature (#177)
- [827a92b6](https://github.com/kubedb/mysql/commit/827a92b6) Mount mysql config dir and tmp dir as emptydir (#166)
- [2a84ed08](https://github.com/kubedb/mysql/commit/2a84ed08) Enable subresource for MySQL crd. (#175)
- [bc8ec773](https://github.com/kubedb/mysql/commit/bc8ec773) Update kubernetes client-go to 1.16.3 (#174)
- [014f6b0b](https://github.com/kubedb/mysql/commit/014f6b0b) Matrix tests for github actions (#172)
- [68f427db](https://github.com/kubedb/mysql/commit/68f427db) Fix default make command
- [76dc7d7b](https://github.com/kubedb/mysql/commit/76dc7d7b) Use charts to install operator (#173)
- [5ff41dc1](https://github.com/kubedb/mysql/commit/5ff41dc1) Add add-license make target
- [132b2a0e](https://github.com/kubedb/mysql/commit/132b2a0e) Add license header to files (#171)
- [aab6050e](https://github.com/kubedb/mysql/commit/aab6050e) Fix linter errors. (#169)
- [35043a15](https://github.com/kubedb/mysql/commit/35043a15) Enable make ci (#168)
- [e452bb4b](https://github.com/kubedb/mysql/commit/e452bb4b) Remove EnableStatusSubresource (#167)
- [28794570](https://github.com/kubedb/mysql/commit/28794570) Run e2e tests using GitHub actions (#164)
- [af3b284b](https://github.com/kubedb/mysql/commit/af3b284b) Validate DBVersionSpecs and fixed broken build (#165)
- [e4963763](https://github.com/kubedb/mysql/commit/e4963763) Update go.yml
- [a808e508](https://github.com/kubedb/mysql/commit/a808e508) Enable GitHub actions
- [6fe5dd42](https://github.com/kubedb/mysql/commit/6fe5dd42) Update changelog



## [kubedb/mysql-replication-mode-detector](https://github.com/kubedb/mysql-replication-mode-detector)

### [v0.1.0-beta.1](https://github.com/kubedb/mysql-replication-mode-detector/releases/tag/v0.1.0-beta.1)

- [3e62838](https://github.com/kubedb/mysql-replication-mode-detector/commit/3e62838) Prepare for release v0.1.0-beta.1 (#9)
- [e54c4c0](https://github.com/kubedb/mysql-replication-mode-detector/commit/e54c4c0) Update License (#7)
- [e071b02](https://github.com/kubedb/mysql-replication-mode-detector/commit/e071b02) Update to Kubernetes v1.18.3 (#6)
- [8992bcb](https://github.com/kubedb/mysql-replication-mode-detector/commit/8992bcb) Update update-release-tracker.sh
- [acc1038](https://github.com/kubedb/mysql-replication-mode-detector/commit/acc1038) Add script to update release tracker on pr merge (#5)
- [706b5b0](https://github.com/kubedb/mysql-replication-mode-detector/commit/706b5b0) Update .kodiak.toml
- [4e52c03](https://github.com/kubedb/mysql-replication-mode-detector/commit/4e52c03) Update to Kubernetes v1.18.3 (#4)
- [adb05ae](https://github.com/kubedb/mysql-replication-mode-detector/commit/adb05ae) Merge branch 'master' into gomod-refresher-1591418508
- [3a99f80](https://github.com/kubedb/mysql-replication-mode-detector/commit/3a99f80) Create .kodiak.toml
- [6289807](https://github.com/kubedb/mysql-replication-mode-detector/commit/6289807) Update to Kubernetes v1.18.3
- [1dd24be](https://github.com/kubedb/mysql-replication-mode-detector/commit/1dd24be) Update to Kubernetes v1.18.3 (#3)
- [6d02366](https://github.com/kubedb/mysql-replication-mode-detector/commit/6d02366) Update Makefile and CI configuration (#2)
- [fc95884](https://github.com/kubedb/mysql-replication-mode-detector/commit/fc95884) Add primary role labeler controller (#1)
- [99dfb12](https://github.com/kubedb/mysql-replication-mode-detector/commit/99dfb12) add readme.md



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.14.0-beta.1](https://github.com/kubedb/operator/releases/tag/v0.14.0-beta.1)

- [a2bba612](https://github.com/kubedb/operator/commit/a2bba612) Prepare for release v0.14.0-beta.1 (#262)
- [22bc85ec](https://github.com/kubedb/operator/commit/22bc85ec) Allow customizing chart registry (#261)
- [52cc1dc7](https://github.com/kubedb/operator/commit/52cc1dc7) Update for release Stash@v2020.07.09-beta.0 (#260)
- [2e8b709f](https://github.com/kubedb/operator/commit/2e8b709f) Update for release Stash@v2020.07.08-beta.0 (#259)
- [7b58b548](https://github.com/kubedb/operator/commit/7b58b548) Update License (#258)
- [d4cd1a93](https://github.com/kubedb/operator/commit/d4cd1a93) Update to Kubernetes v1.18.3 (#256)
- [f6091845](https://github.com/kubedb/operator/commit/f6091845) Update ci.yml
- [5324d2b6](https://github.com/kubedb/operator/commit/5324d2b6) Update ci.yml
- [c888d7fd](https://github.com/kubedb/operator/commit/c888d7fd) Add workflow to update docs (#255)
- [ba843e17](https://github.com/kubedb/operator/commit/ba843e17) Update update-release-tracker.sh
- [b93c5ab4](https://github.com/kubedb/operator/commit/b93c5ab4) Update update-release-tracker.sh
- [6b8d2149](https://github.com/kubedb/operator/commit/6b8d2149) Add script to update release tracker on pr merge (#254)
- [bb1290dc](https://github.com/kubedb/operator/commit/bb1290dc) Update .kodiak.toml
- [9bb85c3b](https://github.com/kubedb/operator/commit/9bb85c3b) Register validator & mutators for all supported dbs (#253)
- [1a524d9c](https://github.com/kubedb/operator/commit/1a524d9c) Various fixes (#252)
- [4860f2a7](https://github.com/kubedb/operator/commit/4860f2a7) Update to Kubernetes v1.18.3 (#251)
- [1a163c6a](https://github.com/kubedb/operator/commit/1a163c6a) Create .kodiak.toml
- [1eda36b9](https://github.com/kubedb/operator/commit/1eda36b9) Update to Kubernetes v1.18.3 (#247)
- [77b8b858](https://github.com/kubedb/operator/commit/77b8b858) Update Enterprise operator tag (#246)
- [96ca876e](https://github.com/kubedb/operator/commit/96ca876e) Revendor kubedb.dev/apimachinery@master (#245)
- [43a3a7f1](https://github.com/kubedb/operator/commit/43a3a7f1) Use recommended kubernetes app labels
- [1ae7045f](https://github.com/kubedb/operator/commit/1ae7045f) Update crazy-max/ghaction-docker-buildx flag
- [f25034ef](https://github.com/kubedb/operator/commit/f25034ef) Trigger the workflow on push or pull request
- [ba486319](https://github.com/kubedb/operator/commit/ba486319) Update readme (#244)
- [5f7191f4](https://github.com/kubedb/operator/commit/5f7191f4) Update CHANGELOG.md
- [5b14af4b](https://github.com/kubedb/operator/commit/5b14af4b) Add license scan report and status (#241)
- [9848932b](https://github.com/kubedb/operator/commit/9848932b) Pass the topology object to common controller
- [90d1c873](https://github.com/kubedb/operator/commit/90d1c873) Initialize topology for MonogDB webhooks (#243)
- [8ecb87c8](https://github.com/kubedb/operator/commit/8ecb87c8) Fix nil pointer exception (#242)
- [b12c3392](https://github.com/kubedb/operator/commit/b12c3392) Update operator dependencies (#237)
- [f714bb1b](https://github.com/kubedb/operator/commit/f714bb1b) Always create RBAC resources (#238)
- [f43a588e](https://github.com/kubedb/operator/commit/f43a588e) Use Go 1.13 in CI
- [e8ab3580](https://github.com/kubedb/operator/commit/e8ab3580) Update client-go to kubernetes-1.16.3 (#239)
- [1dc84a67](https://github.com/kubedb/operator/commit/1dc84a67) Update CI badge
- [d9d1cc0a](https://github.com/kubedb/operator/commit/d9d1cc0a) Bundle PgBouncer operator (#236)
- [720303c1](https://github.com/kubedb/operator/commit/720303c1) Fix linter errors (#235)
- [4c53a71f](https://github.com/kubedb/operator/commit/4c53a71f) Update go.yml
- [e65fc457](https://github.com/kubedb/operator/commit/e65fc457) Enable GitHub actions
- [2dcb0d6d](https://github.com/kubedb/operator/commit/2dcb0d6d) Update changelog



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.1.0-beta.1](https://github.com/kubedb/percona-xtradb/releases/tag/v0.1.0-beta.1)

- [28b9fc0f](https://github.com/kubedb/percona-xtradb/commit/28b9fc0f) Prepare for release v0.1.0-beta.1 (#41)
- [fb4f5444](https://github.com/kubedb/percona-xtradb/commit/fb4f5444) Update for release Stash@v2020.07.09-beta.0 (#39)
- [ad221aa2](https://github.com/kubedb/percona-xtradb/commit/ad221aa2) include Makefile.env
- [841ec855](https://github.com/kubedb/percona-xtradb/commit/841ec855) Allow customizing chart registry (#38)
- [bb608980](https://github.com/kubedb/percona-xtradb/commit/bb608980) Update License (#37)
- [cf8cd2fa](https://github.com/kubedb/percona-xtradb/commit/cf8cd2fa) Update for release Stash@v2020.07.08-beta.0 (#36)
- [7b28c4b9](https://github.com/kubedb/percona-xtradb/commit/7b28c4b9) Update to Kubernetes v1.18.3 (#35)
- [848ff94a](https://github.com/kubedb/percona-xtradb/commit/848ff94a) Update ci.yml
- [d124dd6a](https://github.com/kubedb/percona-xtradb/commit/d124dd6a) Load stash version from .env file for make (#34)
- [1de40e1d](https://github.com/kubedb/percona-xtradb/commit/1de40e1d) Update update-release-tracker.sh
- [7a4503be](https://github.com/kubedb/percona-xtradb/commit/7a4503be) Update update-release-tracker.sh
- [ad0dfaf8](https://github.com/kubedb/percona-xtradb/commit/ad0dfaf8) Add script to update release tracker on pr merge (#33)
- [aaca6bd9](https://github.com/kubedb/percona-xtradb/commit/aaca6bd9) Update .kodiak.toml
- [9a495724](https://github.com/kubedb/percona-xtradb/commit/9a495724) Various fixes (#32)
- [9b6c9a53](https://github.com/kubedb/percona-xtradb/commit/9b6c9a53) Update to Kubernetes v1.18.3 (#31)
- [67912547](https://github.com/kubedb/percona-xtradb/commit/67912547) Update to Kubernetes v1.18.3
- [fc8ce4cc](https://github.com/kubedb/percona-xtradb/commit/fc8ce4cc) Create .kodiak.toml
- [8aba5ef2](https://github.com/kubedb/percona-xtradb/commit/8aba5ef2) Use CRD v1 for Kubernetes >= 1.16 (#30)
- [e81d2b4c](https://github.com/kubedb/percona-xtradb/commit/e81d2b4c) Update to Kubernetes v1.18.3 (#29)
- [2a32730a](https://github.com/kubedb/percona-xtradb/commit/2a32730a) Fix e2e tests (#28)
- [a79626d9](https://github.com/kubedb/percona-xtradb/commit/a79626d9) Update stash install commands
- [52fc2059](https://github.com/kubedb/percona-xtradb/commit/52fc2059) Use recommended kubernetes app labels (#27)
- [93dc10ec](https://github.com/kubedb/percona-xtradb/commit/93dc10ec) Update crazy-max/ghaction-docker-buildx flag
- [ce5717e2](https://github.com/kubedb/percona-xtradb/commit/ce5717e2) Revendor kubedb.dev/apimachinery@master (#26)
- [c1ca649d](https://github.com/kubedb/percona-xtradb/commit/c1ca649d) Pass annotations from CRD to AppBinding (#25)
- [f327cc01](https://github.com/kubedb/percona-xtradb/commit/f327cc01) Trigger the workflow on push or pull request
- [02432393](https://github.com/kubedb/percona-xtradb/commit/02432393) Update CHANGELOG.md
- [a89dbc55](https://github.com/kubedb/percona-xtradb/commit/a89dbc55) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#24)
- [e69742de](https://github.com/kubedb/percona-xtradb/commit/e69742de) Update for percona-xtradb standalone restoresession (#23)
- [958877a1](https://github.com/kubedb/percona-xtradb/commit/958877a1) Various fixes (#21)
- [fb0d7a35](https://github.com/kubedb/percona-xtradb/commit/fb0d7a35) Update kubernetes client-go to 1.16.3 (#20)
- [293fe9a4](https://github.com/kubedb/percona-xtradb/commit/293fe9a4) Fix default make command
- [39358e3b](https://github.com/kubedb/percona-xtradb/commit/39358e3b) Use charts to install operator (#19)
- [6c5b3395](https://github.com/kubedb/percona-xtradb/commit/6c5b3395) Several fixes and update tests (#18)
- [84ff139f](https://github.com/kubedb/percona-xtradb/commit/84ff139f) Various Makefile improvements (#16)
- [e2737f65](https://github.com/kubedb/percona-xtradb/commit/e2737f65) Remove EnableStatusSubresource (#17)
- [fb886b07](https://github.com/kubedb/percona-xtradb/commit/fb886b07) Run e2e tests using GitHub actions (#12)
- [35b155d9](https://github.com/kubedb/percona-xtradb/commit/35b155d9) Validate DBVersionSpecs and fixed broken build (#15)
- [67794bd9](https://github.com/kubedb/percona-xtradb/commit/67794bd9) Update go.yml
- [f7666354](https://github.com/kubedb/percona-xtradb/commit/f7666354) Various changes for Percona XtraDB (#13)
- [ceb7ba67](https://github.com/kubedb/percona-xtradb/commit/ceb7ba67) Enable GitHub actions
- [f5a112af](https://github.com/kubedb/percona-xtradb/commit/f5a112af) Refactor for ProxySQL Integration (#11)
- [26602049](https://github.com/kubedb/percona-xtradb/commit/26602049) Revendor
- [71957d40](https://github.com/kubedb/percona-xtradb/commit/71957d40) Rename from perconaxtradb to percona-xtradb (#10)
- [b526ccd8](https://github.com/kubedb/percona-xtradb/commit/b526ccd8) Set database version in AppBinding (#7)
- [336e7203](https://github.com/kubedb/percona-xtradb/commit/336e7203) Percona XtraDB Cluster support (#9)
- [71a42f7a](https://github.com/kubedb/percona-xtradb/commit/71a42f7a) Don't set annotation to AppBinding (#8)
- [282298cb](https://github.com/kubedb/percona-xtradb/commit/282298cb) Fix UpsertDatabaseAnnotation() function (#4)
- [2ab9dddf](https://github.com/kubedb/percona-xtradb/commit/2ab9dddf) Add license header to Makefiles (#6)
- [df135c08](https://github.com/kubedb/percona-xtradb/commit/df135c08) Add install, uninstall and purge command in Makefile (#3)
- [73d3a845](https://github.com/kubedb/percona-xtradb/commit/73d3a845) Update .gitignore
- [59a4e754](https://github.com/kubedb/percona-xtradb/commit/59a4e754) Add Makefile (#2)
- [f3551ddc](https://github.com/kubedb/percona-xtradb/commit/f3551ddc) Rename package path (#1)
- [56a241d6](https://github.com/kubedb/percona-xtradb/commit/56a241d6) Use explicit IP whitelist instead of automatic IP whitelist (#151)
- [9f0b5ca3](https://github.com/kubedb/percona-xtradb/commit/9f0b5ca3) Update to k8s 1.14.0 client libraries using go.mod (#147)
- [73ad7c30](https://github.com/kubedb/percona-xtradb/commit/73ad7c30) Update changelog
- [ccc36b5c](https://github.com/kubedb/percona-xtradb/commit/ccc36b5c) Update README.md
- [9769e8e1](https://github.com/kubedb/percona-xtradb/commit/9769e8e1) Start next dev cycle
- [a3fa468a](https://github.com/kubedb/percona-xtradb/commit/a3fa468a) Prepare release 0.5.0
- [6d8862de](https://github.com/kubedb/percona-xtradb/commit/6d8862de) Mysql Group Replication tests (#146)
- [49544e55](https://github.com/kubedb/percona-xtradb/commit/49544e55) Mysql Group Replication (#144)
- [a85d4b44](https://github.com/kubedb/percona-xtradb/commit/a85d4b44) Revendor dependencies
- [9c538460](https://github.com/kubedb/percona-xtradb/commit/9c538460) Changed Role to exclude psp without name (#143)
- [6cace93b](https://github.com/kubedb/percona-xtradb/commit/6cace93b) Modify mutator validator names (#142)
- [da0c19b9](https://github.com/kubedb/percona-xtradb/commit/da0c19b9) Update changelog
- [b79c80d6](https://github.com/kubedb/percona-xtradb/commit/b79c80d6) Start next dev cycle
- [838d9459](https://github.com/kubedb/percona-xtradb/commit/838d9459) Prepare release 0.4.0
- [bf0f2c14](https://github.com/kubedb/percona-xtradb/commit/bf0f2c14) Added PSP names and init container image in testing framework (#141)
- [3d227570](https://github.com/kubedb/percona-xtradb/commit/3d227570) Added PSP support for mySQL (#137)
- [7b766657](https://github.com/kubedb/percona-xtradb/commit/7b766657) Don't inherit app.kubernetes.io labels from CRD into offshoots (#140)
- [29e23470](https://github.com/kubedb/percona-xtradb/commit/29e23470) Support for init container (#139)
- [3e1556f6](https://github.com/kubedb/percona-xtradb/commit/3e1556f6) Add role label to stats service (#138)
- [ee078af9](https://github.com/kubedb/percona-xtradb/commit/ee078af9) Update changelog
- [978f1139](https://github.com/kubedb/percona-xtradb/commit/978f1139) Update Kubernetes client libraries to 1.13.0 release (#136)
- [821f23d1](https://github.com/kubedb/percona-xtradb/commit/821f23d1) Start next dev cycle
- [678b26aa](https://github.com/kubedb/percona-xtradb/commit/678b26aa) Prepare release 0.3.0
- [40ad7a23](https://github.com/kubedb/percona-xtradb/commit/40ad7a23) Initial RBAC support: create and use K8s service account for MySQL (#134)
- [98f03387](https://github.com/kubedb/percona-xtradb/commit/98f03387) Revendor dependencies (#135)
- [dfe92615](https://github.com/kubedb/percona-xtradb/commit/dfe92615) Revendor dependencies : Retry Failed Scheduler Snapshot (#133)
- [71f8a350](https://github.com/kubedb/percona-xtradb/commit/71f8a350) Added ephemeral StorageType support (#132)
- [0a6b6e46](https://github.com/kubedb/percona-xtradb/commit/0a6b6e46) Added support of MySQL 8.0.14 (#131)
- [99e57a9e](https://github.com/kubedb/percona-xtradb/commit/99e57a9e) Use PVC spec from snapshot if provided (#130)
- [61497be6](https://github.com/kubedb/percona-xtradb/commit/61497be6) Revendored and updated tests for 'Prevent prefix matching of multiple snapshots' (#129)
- [7eafe088](https://github.com/kubedb/percona-xtradb/commit/7eafe088) Add certificate health checker (#128)
- [973ec416](https://github.com/kubedb/percona-xtradb/commit/973ec416) Update E2E test: Env update is not restricted anymore (#127)
- [339975ff](https://github.com/kubedb/percona-xtradb/commit/339975ff) Fix AppBinding (#126)
- [62050a72](https://github.com/kubedb/percona-xtradb/commit/62050a72) Update changelog
- [2d454043](https://github.com/kubedb/percona-xtradb/commit/2d454043) Prepare release 0.2.0
- [6941ea59](https://github.com/kubedb/percona-xtradb/commit/6941ea59) Reuse event recorder (#125)
- [b77e66c4](https://github.com/kubedb/percona-xtradb/commit/b77e66c4) OSM binary upgraded in mysql-tools (#123)
- [c9228086](https://github.com/kubedb/percona-xtradb/commit/c9228086) Revendor dependencies (#124)
- [97837120](https://github.com/kubedb/percona-xtradb/commit/97837120) Test for faulty snapshot (#122)
- [c3e995b6](https://github.com/kubedb/percona-xtradb/commit/c3e995b6) Start next dev cycle
- [8a4f3b13](https://github.com/kubedb/percona-xtradb/commit/8a4f3b13) Prepare release 0.2.0-rc.2
- [79942191](https://github.com/kubedb/percona-xtradb/commit/79942191) Upgrade database secret keys (#121)
- [1747fdf5](https://github.com/kubedb/percona-xtradb/commit/1747fdf5) Ignore mutation of fields to default values during update (#120)
- [d902d588](https://github.com/kubedb/percona-xtradb/commit/d902d588) Support configuration options for exporter sidecar (#119)
- [dd7c3f44](https://github.com/kubedb/percona-xtradb/commit/dd7c3f44) Use flags.DumpAll (#118)
- [bc1ef05b](https://github.com/kubedb/percona-xtradb/commit/bc1ef05b) Start next dev cycle
- [9d33c1a0](https://github.com/kubedb/percona-xtradb/commit/9d33c1a0) Prepare release 0.2.0-rc.1
- [b076e141](https://github.com/kubedb/percona-xtradb/commit/b076e141) Apply cleanup (#117)
- [7dc5641f](https://github.com/kubedb/percona-xtradb/commit/7dc5641f) Set periodic analytics (#116)
- [90ea6acc](https://github.com/kubedb/percona-xtradb/commit/90ea6acc) Introduce AppBinding support (#115)
- [a882d76a](https://github.com/kubedb/percona-xtradb/commit/a882d76a) Fix Analytics (#114)
- [0961009c](https://github.com/kubedb/percona-xtradb/commit/0961009c) Error out from cron job for deprecated dbversion (#113)
- [da1f4e27](https://github.com/kubedb/percona-xtradb/commit/da1f4e27) Add CRDs without observation when operator starts (#112)
- [0a754d2f](https://github.com/kubedb/percona-xtradb/commit/0a754d2f) Update changelog
- [b09bc6e1](https://github.com/kubedb/percona-xtradb/commit/b09bc6e1) Start next dev cycle
- [0d467ccb](https://github.com/kubedb/percona-xtradb/commit/0d467ccb) Prepare release 0.2.0-rc.0
- [c757007a](https://github.com/kubedb/percona-xtradb/commit/c757007a) Merge commit 'cc6607a3589a79a5e61bb198d370ea0ae30b9d09'
- [ddfe4be1](https://github.com/kubedb/percona-xtradb/commit/ddfe4be1) Support custom user passowrd for backup (#111)
- [8c84ba20](https://github.com/kubedb/percona-xtradb/commit/8c84ba20) Support providing resources for monitoring container (#110)
- [7bcfbc48](https://github.com/kubedb/percona-xtradb/commit/7bcfbc48) Update kubernetes client libraries to 1.12.0 (#109)
- [145bba2b](https://github.com/kubedb/percona-xtradb/commit/145bba2b) Add validation webhook xray (#108)
- [6da1887f](https://github.com/kubedb/percona-xtradb/commit/6da1887f) Various Fixes (#107)
- [111519e9](https://github.com/kubedb/percona-xtradb/commit/111519e9) Merge ports from service template (#105)
- [38147ef1](https://github.com/kubedb/percona-xtradb/commit/38147ef1) Replace doNotPause with TerminationPolicy = DoNotTerminate (#104)
- [e28ebc47](https://github.com/kubedb/percona-xtradb/commit/e28ebc47) Pass resources to NamespaceValidator (#103)
- [aed12bf5](https://github.com/kubedb/percona-xtradb/commit/aed12bf5) Various fixes (#102)
- [3d372ef6](https://github.com/kubedb/percona-xtradb/commit/3d372ef6) Support Livecycle hook and container probes (#101)
- [b6ef6887](https://github.com/kubedb/percona-xtradb/commit/b6ef6887) Check if Kubernetes version is supported before running operator (#100)
- [d89e7783](https://github.com/kubedb/percona-xtradb/commit/d89e7783) Update package alias (#99)
- [f0b44b3a](https://github.com/kubedb/percona-xtradb/commit/f0b44b3a) Start next dev cycle
- [a79ff03b](https://github.com/kubedb/percona-xtradb/commit/a79ff03b) Prepare release 0.2.0-beta.1
- [0d8d3cca](https://github.com/kubedb/percona-xtradb/commit/0d8d3cca) Revendor api (#98)
- [2f850243](https://github.com/kubedb/percona-xtradb/commit/2f850243) Fix tests (#97)
- [4ced0bfe](https://github.com/kubedb/percona-xtradb/commit/4ced0bfe) Revendor api for catalog apigroup (#96)
- [e7695400](https://github.com/kubedb/percona-xtradb/commit/e7695400) Update chanelog
- [8e358aea](https://github.com/kubedb/percona-xtradb/commit/8e358aea) Use --pull flag with docker build (#20) (#95)
- [d2a97d90](https://github.com/kubedb/percona-xtradb/commit/d2a97d90) Merge commit '16c769ee4686576f172a6b79a10d25bfd79ca4a4'
- [d1fe8a8a](https://github.com/kubedb/percona-xtradb/commit/d1fe8a8a) Start next dev cycle
- [04eb9bb5](https://github.com/kubedb/percona-xtradb/commit/04eb9bb5) Prepare release 0.2.0-beta.0
- [9dfea960](https://github.com/kubedb/percona-xtradb/commit/9dfea960) Pass extra args to tools.sh (#93)
- [47dd3cad](https://github.com/kubedb/percona-xtradb/commit/47dd3cad) Don't try to wipe out Snapshot data for Local backend (#92)
- [9c4d485b](https://github.com/kubedb/percona-xtradb/commit/9c4d485b) Add missing alt-tag docker folder mysql-tools images (#91)
- [be72f784](https://github.com/kubedb/percona-xtradb/commit/be72f784) Use suffix for updated DBImage & Stop working for deprecated *Versions (#90)
- [05c8f14d](https://github.com/kubedb/percona-xtradb/commit/05c8f14d) Search used secrets within same namespace of DB object (#89)
- [0d94c946](https://github.com/kubedb/percona-xtradb/commit/0d94c946) Support Termination Policy (#88)
- [8775ddf7](https://github.com/kubedb/percona-xtradb/commit/8775ddf7) Update builddeps.sh
- [796c93da](https://github.com/kubedb/percona-xtradb/commit/796c93da) Revendor k8s.io/apiserver (#87)
- [5a1e3f57](https://github.com/kubedb/percona-xtradb/commit/5a1e3f57) Revendor kubernetes-1.11.3 (#86)
- [809a3c49](https://github.com/kubedb/percona-xtradb/commit/809a3c49) Support UpdateStrategy (#84)
- [372c52ef](https://github.com/kubedb/percona-xtradb/commit/372c52ef) Add TerminationPolicy for databases (#83)
- [c01b55e8](https://github.com/kubedb/percona-xtradb/commit/c01b55e8) Revendor api (#82)
- [5e196b95](https://github.com/kubedb/percona-xtradb/commit/5e196b95) Use IntHash as status.observedGeneration (#81)
- [2da3bb1b](https://github.com/kubedb/percona-xtradb/commit/2da3bb1b) fix github status (#80)
- [121d0a98](https://github.com/kubedb/percona-xtradb/commit/121d0a98) Update pipeline (#79)
- [532e3137](https://github.com/kubedb/percona-xtradb/commit/532e3137) Fix E2E test for minikube (#78)
- [0f107815](https://github.com/kubedb/percona-xtradb/commit/0f107815) Update pipeline (#77)
- [851679e2](https://github.com/kubedb/percona-xtradb/commit/851679e2) Migrate MySQL (#75)
- [0b997855](https://github.com/kubedb/percona-xtradb/commit/0b997855) Use official exporter image (#74)
- [702d5736](https://github.com/kubedb/percona-xtradb/commit/702d5736) Fix uninstall for concourse (#70)
- [9ee88bd2](https://github.com/kubedb/percona-xtradb/commit/9ee88bd2) Update status.ObservedGeneration for failure phase (#73)
- [559cdb6a](https://github.com/kubedb/percona-xtradb/commit/559cdb6a) Keep track of ObservedGenerationHash (#72)
- [61c8b898](https://github.com/kubedb/percona-xtradb/commit/61c8b898) Use NewObservableHandler (#71)
- [421274dc](https://github.com/kubedb/percona-xtradb/commit/421274dc) Merge commit '887037c7e36289e3135dda99346fccc7e2ce303b'
- [6a41d9bc](https://github.com/kubedb/percona-xtradb/commit/6a41d9bc) Fix uninstall for concourse (#69)
- [f1af09db](https://github.com/kubedb/percona-xtradb/commit/f1af09db) Update README.md
- [bf3f1823](https://github.com/kubedb/percona-xtradb/commit/bf3f1823) Revise immutable spec fields (#68)
- [26adec3b](https://github.com/kubedb/percona-xtradb/commit/26adec3b) Merge commit '5f83049fc01dc1d0709ac0014d6f3a0f74a39417'
- [31a97820](https://github.com/kubedb/percona-xtradb/commit/31a97820) Support passing args via PodTemplate (#67)
- [60f4ee23](https://github.com/kubedb/percona-xtradb/commit/60f4ee23) Introduce storageType : ephemeral (#66)
- [bfd3fcd6](https://github.com/kubedb/percona-xtradb/commit/bfd3fcd6) Add support for running tests on cncf cluster (#63)
- [fba47b19](https://github.com/kubedb/percona-xtradb/commit/fba47b19) Merge commit 'e010cbb302c8d59d4cf69dd77085b046ff423b78'
- [6be96ce0](https://github.com/kubedb/percona-xtradb/commit/6be96ce0) Revendor api (#65)
- [0f629ab3](https://github.com/kubedb/percona-xtradb/commit/0f629ab3) Keep track of observedGeneration in status (#64)
- [c9a9596f](https://github.com/kubedb/percona-xtradb/commit/c9a9596f) Separate StatsService for monitoring (#62)
- [62854641](https://github.com/kubedb/percona-xtradb/commit/62854641) Use MySQLVersion for MySQL images (#61)
- [3c170c56](https://github.com/kubedb/percona-xtradb/commit/3c170c56) Use updated crd spec (#60)
- [873c285e](https://github.com/kubedb/percona-xtradb/commit/873c285e) Rename OffshootLabels to OffshootSelectors (#59)
- [2fd02169](https://github.com/kubedb/percona-xtradb/commit/2fd02169) Revendor api (#58)
- [a127d6cd](https://github.com/kubedb/percona-xtradb/commit/a127d6cd) Use kmodules monitoring and objectstore api (#57)
- [2f79a038](https://github.com/kubedb/percona-xtradb/commit/2f79a038) Support custom configuration (#52)
- [49c67f00](https://github.com/kubedb/percona-xtradb/commit/49c67f00) Merge commit '44e6d4985d93556e39ddcc4677ada5437fc5be64'
- [fb28bc6c](https://github.com/kubedb/percona-xtradb/commit/fb28bc6c) Refactor concourse scripts (#56)
- [4de4ced1](https://github.com/kubedb/percona-xtradb/commit/4de4ced1) Fix command `./hack/make.py test e2e` (#55)
- [3082123e](https://github.com/kubedb/percona-xtradb/commit/3082123e) Set generated binary name to my-operator (#54)
- [5698f314](https://github.com/kubedb/percona-xtradb/commit/5698f314) Don't add admission/v1beta1 group as a prioritized version (#53)
- [696135d5](https://github.com/kubedb/percona-xtradb/commit/696135d5) Fix travis build (#48)
- [c519ef89](https://github.com/kubedb/percona-xtradb/commit/c519ef89) Format shell script (#51)
- [c93e2f40](https://github.com/kubedb/percona-xtradb/commit/c93e2f40) Enable status subresource for crds (#50)
- [edd951ca](https://github.com/kubedb/percona-xtradb/commit/edd951ca) Update client-go to v8.0.0 (#49)
- [520597a6](https://github.com/kubedb/percona-xtradb/commit/520597a6) Merge commit '71850e2c90cda8fc588b7dedb340edf3d316baea'
- [f1549e95](https://github.com/kubedb/percona-xtradb/commit/f1549e95) Support ENV variables in CRDs (#46)
- [67f37780](https://github.com/kubedb/percona-xtradb/commit/67f37780) Updated osm version to 0.7.1 (#47)
- [10e309c0](https://github.com/kubedb/percona-xtradb/commit/10e309c0) Prepare release 0.1.0
- [62a8fbbd](https://github.com/kubedb/percona-xtradb/commit/62a8fbbd) Fixed missing error return (#45)
- [8c05bb83](https://github.com/kubedb/percona-xtradb/commit/8c05bb83) Revendor dependencies (#44)
- [ca811a2e](https://github.com/kubedb/percona-xtradb/commit/ca811a2e) Fix release script (#43)
- [b79541f6](https://github.com/kubedb/percona-xtradb/commit/b79541f6) Add changelog (#42)
- [a2d13c82](https://github.com/kubedb/percona-xtradb/commit/a2d13c82) Concourse (#41)
- [95b2186e](https://github.com/kubedb/percona-xtradb/commit/95b2186e) Fixed kubeconfig plugin for Cloud Providers && Storage is required for MySQL (#40)
- [37762093](https://github.com/kubedb/percona-xtradb/commit/37762093) Refactored E2E testing to support E2E testing with admission webhook in cloud (#38)
- [b6fe72ca](https://github.com/kubedb/percona-xtradb/commit/b6fe72ca) Remove lost+found directory before initializing mysql (#39)
- [18ebb959](https://github.com/kubedb/percona-xtradb/commit/18ebb959) Skip delete requests for empty resources (#37)
- [eeb7add0](https://github.com/kubedb/percona-xtradb/commit/eeb7add0) Don't panic if admission options is nil (#36)
- [ccb59db0](https://github.com/kubedb/percona-xtradb/commit/ccb59db0) Disable admission controllers for webhook server (#35)
- [b1c6c149](https://github.com/kubedb/percona-xtradb/commit/b1c6c149) Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 (#34)
- [b1890f7c](https://github.com/kubedb/percona-xtradb/commit/b1890f7c) Update client-go to 7.0.0 (#33)
- [08c81726](https://github.com/kubedb/percona-xtradb/commit/08c81726) Added update script for mysql-tools:8 (#32)
- [4bbe6c9f](https://github.com/kubedb/percona-xtradb/commit/4bbe6c9f) Added support of mysql:5.7 (#31)
- [e657f512](https://github.com/kubedb/percona-xtradb/commit/e657f512) Add support for one informer and N-eventHandler for snapshot, dromantDB and Job (#30)
- [bbcd48d6](https://github.com/kubedb/percona-xtradb/commit/bbcd48d6) Use metrics from kube apiserver (#29)
- [1687e197](https://github.com/kubedb/percona-xtradb/commit/1687e197) Bundle webhook server and Use SharedInformerFactory (#28)
- [cd0efc00](https://github.com/kubedb/percona-xtradb/commit/cd0efc00) Move MySQL AdmissionWebhook packages into MySQL repository (#27)
- [46065e18](https://github.com/kubedb/percona-xtradb/commit/46065e18) Use mysql:8.0.3 image as mysql:8.0 (#26)
- [1b73529f](https://github.com/kubedb/percona-xtradb/commit/1b73529f) Update README.md
- [62eaa397](https://github.com/kubedb/percona-xtradb/commit/62eaa397) Update README.md
- [c53704c7](https://github.com/kubedb/percona-xtradb/commit/c53704c7) Remove Docker pull count
- [b9ec877e](https://github.com/kubedb/percona-xtradb/commit/b9ec877e) Add travis yaml (#25)
- [ade3571c](https://github.com/kubedb/percona-xtradb/commit/ade3571c) Start next dev cycle
- [b4b749df](https://github.com/kubedb/percona-xtradb/commit/b4b749df) Prepare release 0.1.0-beta.2
- [4d46d95d](https://github.com/kubedb/percona-xtradb/commit/4d46d95d) Migrating to apps/v1 (#23)
- [5ee1ac8c](https://github.com/kubedb/percona-xtradb/commit/5ee1ac8c) Update validation (#22)
- [dd023c50](https://github.com/kubedb/percona-xtradb/commit/dd023c50)  Fix dormantDB matching: pass same type to Equal method (#21)
- [37a1e4fd](https://github.com/kubedb/percona-xtradb/commit/37a1e4fd) Use official code generator scripts (#20)
- [485d3d7c](https://github.com/kubedb/percona-xtradb/commit/485d3d7c) Fixed dormantdb matching & Raised throttling time & Fixed MySQL version Checking (#19)
- [6db2ae8d](https://github.com/kubedb/percona-xtradb/commit/6db2ae8d) Prepare release 0.1.0-beta.1
- [ebbfec2f](https://github.com/kubedb/percona-xtradb/commit/ebbfec2f) converted to k8s 1.9 & Improved InitSpec in DormantDB & Added support for Job watcher & Improved Tests (#17)
- [a484e0e5](https://github.com/kubedb/percona-xtradb/commit/a484e0e5) Fixed logger, analytics and removed rbac stuff (#16)
- [7aa2d1d2](https://github.com/kubedb/percona-xtradb/commit/7aa2d1d2) Add rbac stuffs for mysql-exporter (#15)
- [078098c8](https://github.com/kubedb/percona-xtradb/commit/078098c8)  Review Mysql docker images and Fixed monitring (#14)
- [6877108a](https://github.com/kubedb/percona-xtradb/commit/6877108a) Update README.md
- [1f84a5da](https://github.com/kubedb/percona-xtradb/commit/1f84a5da) Start next dev cycle
- [2f1e4b7d](https://github.com/kubedb/percona-xtradb/commit/2f1e4b7d) Prepare release 0.1.0-beta.0
- [dce1e88e](https://github.com/kubedb/percona-xtradb/commit/dce1e88e) Add release script
- [60ed55cb](https://github.com/kubedb/percona-xtradb/commit/60ed55cb) Rename ms-operator to my-operator (#13)
- [5451d166](https://github.com/kubedb/percona-xtradb/commit/5451d166) Fix Analytics and pass client-id as ENV to Snapshot Job (#12)
- [788ae178](https://github.com/kubedb/percona-xtradb/commit/788ae178) update docker image validation (#11)
- [c966efd5](https://github.com/kubedb/percona-xtradb/commit/c966efd5) Add docker-registry and WorkQueue  (#10)
- [be340103](https://github.com/kubedb/percona-xtradb/commit/be340103) Set client id for analytics (#9)
- [ca11f683](https://github.com/kubedb/percona-xtradb/commit/ca11f683) Fix CRD Registration (#8)
- [2f95c13d](https://github.com/kubedb/percona-xtradb/commit/2f95c13d) Update issue repo link
- [6fffa713](https://github.com/kubedb/percona-xtradb/commit/6fffa713) Update pkg paths to kubedb org (#7)
- [2d4d5c44](https://github.com/kubedb/percona-xtradb/commit/2d4d5c44) Assign default Prometheus Monitoring Port (#6)
- [a7595613](https://github.com/kubedb/percona-xtradb/commit/a7595613) Add Snapshot Backup, Restore and Backup-Scheduler (#4)
- [17a782c6](https://github.com/kubedb/percona-xtradb/commit/17a782c6) Update Dockerfile
- [e92bfec9](https://github.com/kubedb/percona-xtradb/commit/e92bfec9) Add mysql-util docker image (#5)
- [2a4b25ac](https://github.com/kubedb/percona-xtradb/commit/2a4b25ac) Mysql db - Inititalizing  (#2)
- [cbfbc878](https://github.com/kubedb/percona-xtradb/commit/cbfbc878) Update README.md
- [01cab651](https://github.com/kubedb/percona-xtradb/commit/01cab651) Update README.md
- [0aa81cdf](https://github.com/kubedb/percona-xtradb/commit/0aa81cdf) Use client-go 5.x
- [3de10d7f](https://github.com/kubedb/percona-xtradb/commit/3de10d7f) Update ./hack folder (#3)
- [46f05b1f](https://github.com/kubedb/percona-xtradb/commit/46f05b1f) Add skeleton for mysql (#1)
- [73147dba](https://github.com/kubedb/percona-xtradb/commit/73147dba) Merge commit 'be70502b4993171bbad79d2ff89a9844f1c24caa' as 'hack/libbuild'



## [kubedb/pg-leader-election](https://github.com/kubedb/pg-leader-election)

### [v0.2.0-beta.1](https://github.com/kubedb/pg-leader-election/releases/tag/v0.2.0-beta.1)




## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.1.0-beta.1](https://github.com/kubedb/pgbouncer/releases/tag/v0.1.0-beta.1)

- [bbf810c](https://github.com/kubedb/pgbouncer/commit/bbf810c) Prepare for release v0.1.0-beta.1 (#23)
- [5a6e361](https://github.com/kubedb/pgbouncer/commit/5a6e361) include Makefile.env (#22)
- [2d52d66](https://github.com/kubedb/pgbouncer/commit/2d52d66) Update License (#21)
- [33305d5](https://github.com/kubedb/pgbouncer/commit/33305d5) Update to Kubernetes v1.18.3 (#20)
- [b443a55](https://github.com/kubedb/pgbouncer/commit/b443a55) Update ci.yml
- [d3bedc9](https://github.com/kubedb/pgbouncer/commit/d3bedc9) Update update-release-tracker.sh
- [d9100ec](https://github.com/kubedb/pgbouncer/commit/d9100ec) Update update-release-tracker.sh
- [9b86bda](https://github.com/kubedb/pgbouncer/commit/9b86bda) Add script to update release tracker on pr merge (#19)
- [3362cef](https://github.com/kubedb/pgbouncer/commit/3362cef) Update .kodiak.toml
- [11ebebd](https://github.com/kubedb/pgbouncer/commit/11ebebd) Use POSTGRES_TAG v0.14.0-alpha.0
- [dbe95b5](https://github.com/kubedb/pgbouncer/commit/dbe95b5) Various fixes (#18)
- [c50c65d](https://github.com/kubedb/pgbouncer/commit/c50c65d) Update to Kubernetes v1.18.3 (#17)
- [483fa43](https://github.com/kubedb/pgbouncer/commit/483fa43) Update to Kubernetes v1.18.3
- [c0fa8e4](https://github.com/kubedb/pgbouncer/commit/c0fa8e4) Create .kodiak.toml
- [5e33801](https://github.com/kubedb/pgbouncer/commit/5e33801) Use CRD v1 for Kubernetes >= 1.16 (#16)
- [ef7fe47](https://github.com/kubedb/pgbouncer/commit/ef7fe47) Update to Kubernetes v1.18.3 (#15)
- [063339f](https://github.com/kubedb/pgbouncer/commit/063339f) Fix e2e tests (#14)
- [7cd92ba](https://github.com/kubedb/pgbouncer/commit/7cd92ba) Update crazy-max/ghaction-docker-buildx flag
- [e7a47a5](https://github.com/kubedb/pgbouncer/commit/e7a47a5) Revendor kubedb.dev/apimachinery@master (#13)
- [9d00916](https://github.com/kubedb/pgbouncer/commit/9d00916) Use updated operator labels in e2e tests (#12)
- [778924a](https://github.com/kubedb/pgbouncer/commit/778924a) Trigger the workflow on push or pull request
- [77be6b9](https://github.com/kubedb/pgbouncer/commit/77be6b9) Update CHANGELOG.md
- [a9decb9](https://github.com/kubedb/pgbouncer/commit/a9decb9) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#11)
- [cd4d272](https://github.com/kubedb/pgbouncer/commit/cd4d272) Fix build
- [b21b1a1](https://github.com/kubedb/pgbouncer/commit/b21b1a1) Revendor and update enterprise sidecar image (#10)
- [463f7bc](https://github.com/kubedb/pgbouncer/commit/463f7bc) Update Enterprise operator tag (#9)
- [6e01588](https://github.com/kubedb/pgbouncer/commit/6e01588) Use kubedb/installer master branch in CI
- [88b98a4](https://github.com/kubedb/pgbouncer/commit/88b98a4) Update pgbouncer controller (#8)
- [a6b71bc](https://github.com/kubedb/pgbouncer/commit/a6b71bc) Update variable names
- [1a6794b](https://github.com/kubedb/pgbouncer/commit/1a6794b) Fix plain text secret in exporter container of StatefulSet (#5)
- [ab104a9](https://github.com/kubedb/pgbouncer/commit/ab104a9) Update client-go to kubernetes-1.16.3 (#7)
- [68dbb14](https://github.com/kubedb/pgbouncer/commit/68dbb14) Use charts to install operator (#6)
- [30e3e72](https://github.com/kubedb/pgbouncer/commit/30e3e72) Add add-license make target
- [6c1a78a](https://github.com/kubedb/pgbouncer/commit/6c1a78a) Enable e2e tests in GitHub actions (#4)
- [0960f80](https://github.com/kubedb/pgbouncer/commit/0960f80) Initial implementation (#2)
- [a8a9b1d](https://github.com/kubedb/pgbouncer/commit/a8a9b1d) Update go.yml
- [bc3b262](https://github.com/kubedb/pgbouncer/commit/bc3b262) Enable GitHub actions
- [2e33db2](https://github.com/kubedb/pgbouncer/commit/2e33db2) Clone kubedb/postgres repo (#1)
- [45a7cac](https://github.com/kubedb/pgbouncer/commit/45a7cac) Merge commit 'f78de886ed657650438f99574c3b002dd3607497' as 'hack/libbuild'



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.14.0-beta.1](https://github.com/kubedb/postgres/releases/tag/v0.14.0-beta.1)

- [3848a43e](https://github.com/kubedb/postgres/commit/3848a43e) Prepare for release v0.14.0-beta.1 (#325)
- [d4ea0ba7](https://github.com/kubedb/postgres/commit/d4ea0ba7) Update for release Stash@v2020.07.09-beta.0 (#323)
- [6974afda](https://github.com/kubedb/postgres/commit/6974afda) Allow customizing kube namespace for Stash
- [d7d79ea1](https://github.com/kubedb/postgres/commit/d7d79ea1) Allow customizing chart registry (#322)
- [ba0423ac](https://github.com/kubedb/postgres/commit/ba0423ac) Update for release Stash@v2020.07.08-beta.0 (#321)
- [7e855763](https://github.com/kubedb/postgres/commit/7e855763) Update License
- [7bea404a](https://github.com/kubedb/postgres/commit/7bea404a) Update to Kubernetes v1.18.3 (#320)
- [eab0e83f](https://github.com/kubedb/postgres/commit/eab0e83f) Update ci.yml
- [4949f76e](https://github.com/kubedb/postgres/commit/4949f76e) Load stash version from .env file for make (#319)
- [79e9d8d9](https://github.com/kubedb/postgres/commit/79e9d8d9) Update update-release-tracker.sh
- [ca966b7b](https://github.com/kubedb/postgres/commit/ca966b7b) Update update-release-tracker.sh
- [31bbecfe](https://github.com/kubedb/postgres/commit/31bbecfe) Add script to update release tracker on pr merge (#318)
- [540d977f](https://github.com/kubedb/postgres/commit/540d977f) Update .kodiak.toml
- [3e7514a7](https://github.com/kubedb/postgres/commit/3e7514a7) Various fixes (#317)
- [1a5df17c](https://github.com/kubedb/postgres/commit/1a5df17c) Update to Kubernetes v1.18.3 (#315)
- [717cfb3f](https://github.com/kubedb/postgres/commit/717cfb3f) Update to Kubernetes v1.18.3
- [95537169](https://github.com/kubedb/postgres/commit/95537169) Create .kodiak.toml
- [02579005](https://github.com/kubedb/postgres/commit/02579005) Use CRD v1 for Kubernetes >= 1.16 (#314)
- [6ce6deb1](https://github.com/kubedb/postgres/commit/6ce6deb1) Update to Kubernetes v1.18.3 (#313)
- [97f25ba0](https://github.com/kubedb/postgres/commit/97f25ba0) Fix e2e tests (#312)
- [a989c377](https://github.com/kubedb/postgres/commit/a989c377) Update stash install commands
- [6af12596](https://github.com/kubedb/postgres/commit/6af12596) Revendor kubedb.dev/apimachinery@master (#311)
- [9969b064](https://github.com/kubedb/postgres/commit/9969b064) Update crazy-max/ghaction-docker-buildx flag
- [e3360119](https://github.com/kubedb/postgres/commit/e3360119) Use updated operator labels in e2e tests (#309)
- [c183007c](https://github.com/kubedb/postgres/commit/c183007c) Pass annotations from CRD to AppBinding (#310)
- [55581f79](https://github.com/kubedb/postgres/commit/55581f79) Trigger the workflow on push or pull request
- [931b88cf](https://github.com/kubedb/postgres/commit/931b88cf) Update CHANGELOG.md
- [6f481749](https://github.com/kubedb/postgres/commit/6f481749) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#308)
- [15f0611d](https://github.com/kubedb/postgres/commit/15f0611d) Fix error msg to reject halt when termination policy is 'DoNotTerminate'
- [18aba058](https://github.com/kubedb/postgres/commit/18aba058) Change Pause to Halt (#307)
- [7e9b1c69](https://github.com/kubedb/postgres/commit/7e9b1c69) feat: allow changes to nodeSelector (#298)
- [a602faa1](https://github.com/kubedb/postgres/commit/a602faa1) Introduce spec.halted and removed dormant and snapshot crd (#305)
- [cdd384d7](https://github.com/kubedb/postgres/commit/cdd384d7) Moved leader election to kubedb/pg-leader-election (#304)
- [32c41db6](https://github.com/kubedb/postgres/commit/32c41db6) Use stash@v0.9.0-rc.4 release (#306)
- [fa55b472](https://github.com/kubedb/postgres/commit/fa55b472) Make e2e tests stable in github actions (#303)
- [afdc5fda](https://github.com/kubedb/postgres/commit/afdc5fda) Update client-go to kubernetes-1.16.3 (#301)
- [d28eb55a](https://github.com/kubedb/postgres/commit/d28eb55a) Take out postgres docker images and Matrix test (#297)
- [13fee32d](https://github.com/kubedb/postgres/commit/13fee32d) Fix default make command
- [55dfb368](https://github.com/kubedb/postgres/commit/55dfb368) Update catalog values for make install command
- [25f5b79c](https://github.com/kubedb/postgres/commit/25f5b79c) Use charts to install operator (#302)
- [c5a4ed77](https://github.com/kubedb/postgres/commit/c5a4ed77) Add add-license make target
- [aa1d98d0](https://github.com/kubedb/postgres/commit/aa1d98d0) Add license header to files (#296)
- [fd356006](https://github.com/kubedb/postgres/commit/fd356006) Fix E2E testing for github actions (#295)
- [6a3443a7](https://github.com/kubedb/postgres/commit/6a3443a7) Minio and S3 compatible storage fixes (#292)
- [5150cf34](https://github.com/kubedb/postgres/commit/5150cf34) Run e2e tests using GitHub actions (#293)
- [a4a3785b](https://github.com/kubedb/postgres/commit/a4a3785b) Validate DBVersionSpecs and fixed broken build (#294)
- [b171a244](https://github.com/kubedb/postgres/commit/b171a244) Update go.yml
- [1a61bf29](https://github.com/kubedb/postgres/commit/1a61bf29) Enable GitHub actions
- [6b869b15](https://github.com/kubedb/postgres/commit/6b869b15) Update changelog



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.1.0-beta.1](https://github.com/kubedb/proxysql/releases/tag/v0.1.0-beta.1)

- [2ed7d0e8](https://github.com/kubedb/proxysql/commit/2ed7d0e8) Prepare for release v0.1.0-beta.1 (#26)
- [3b5ee481](https://github.com/kubedb/proxysql/commit/3b5ee481) Update for release Stash@v2020.07.09-beta.0 (#25)
- [92b04b33](https://github.com/kubedb/proxysql/commit/92b04b33) include Makefile.env (#24)
- [eace7e26](https://github.com/kubedb/proxysql/commit/eace7e26) Update for release Stash@v2020.07.08-beta.0 (#23)
- [0c647c01](https://github.com/kubedb/proxysql/commit/0c647c01) Update License (#22)
- [3c1b41be](https://github.com/kubedb/proxysql/commit/3c1b41be) Update to Kubernetes v1.18.3 (#21)
- [dfa95bb8](https://github.com/kubedb/proxysql/commit/dfa95bb8) Update ci.yml
- [87390932](https://github.com/kubedb/proxysql/commit/87390932) Update update-release-tracker.sh
- [772a0c6a](https://github.com/kubedb/proxysql/commit/772a0c6a) Update update-release-tracker.sh
- [a3b2ae92](https://github.com/kubedb/proxysql/commit/a3b2ae92) Add script to update release tracker on pr merge (#20)
- [7578cae3](https://github.com/kubedb/proxysql/commit/7578cae3) Update .kodiak.toml
- [4ba876bc](https://github.com/kubedb/proxysql/commit/4ba876bc) Update operator tags
- [399aa60b](https://github.com/kubedb/proxysql/commit/399aa60b) Various fixes (#19)
- [7235b0c5](https://github.com/kubedb/proxysql/commit/7235b0c5) Update to Kubernetes v1.18.3 (#18)
- [427c1f21](https://github.com/kubedb/proxysql/commit/427c1f21) Update to Kubernetes v1.18.3
- [1ac8da55](https://github.com/kubedb/proxysql/commit/1ac8da55) Create .kodiak.toml
- [3243d446](https://github.com/kubedb/proxysql/commit/3243d446) Use CRD v1 for Kubernetes >= 1.16 (#17)
- [4f5bea8d](https://github.com/kubedb/proxysql/commit/4f5bea8d) Update to Kubernetes v1.18.3 (#16)
- [a0d2611a](https://github.com/kubedb/proxysql/commit/a0d2611a) Fix e2e tests (#15)
- [987fbf60](https://github.com/kubedb/proxysql/commit/987fbf60) Update crazy-max/ghaction-docker-buildx flag
- [c2fad78e](https://github.com/kubedb/proxysql/commit/c2fad78e) Use updated operator labels in e2e tests (#14)
- [c5a01db8](https://github.com/kubedb/proxysql/commit/c5a01db8) Revendor kubedb.dev/apimachinery@master (#13)
- [756c8f8f](https://github.com/kubedb/proxysql/commit/756c8f8f) Trigger the workflow on push or pull request
- [fdf84e27](https://github.com/kubedb/proxysql/commit/fdf84e27) Update CHANGELOG.md
- [9075b453](https://github.com/kubedb/proxysql/commit/9075b453) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#12)
- [f4d1c024](https://github.com/kubedb/proxysql/commit/f4d1c024) Matrix Tests on Github Actions (#11)
- [4e021072](https://github.com/kubedb/proxysql/commit/4e021072) Update mount path for custom config (#8)
- [b0922173](https://github.com/kubedb/proxysql/commit/b0922173) Enable ProxySQL monitoring (#6)
- [70be4e67](https://github.com/kubedb/proxysql/commit/70be4e67) ProxySQL test for MySQL (#4)
- [0a444b9e](https://github.com/kubedb/proxysql/commit/0a444b9e) Use charts to install operator (#7)
- [a51fbb51](https://github.com/kubedb/proxysql/commit/a51fbb51) ProxySQL operator for MySQL databases (#2)
- [883fa437](https://github.com/kubedb/proxysql/commit/883fa437) Update go.yml
- [2c0cf51c](https://github.com/kubedb/proxysql/commit/2c0cf51c) Enable GitHub actions
- [52e15cd2](https://github.com/kubedb/proxysql/commit/52e15cd2) percona-xtradb -> proxysql (#1)
- [dc71bffe](https://github.com/kubedb/proxysql/commit/dc71bffe) Revendor
- [71957d40](https://github.com/kubedb/proxysql/commit/71957d40) Rename from perconaxtradb to percona-xtradb (#10)
- [b526ccd8](https://github.com/kubedb/proxysql/commit/b526ccd8) Set database version in AppBinding (#7)
- [336e7203](https://github.com/kubedb/proxysql/commit/336e7203) Percona XtraDB Cluster support (#9)
- [71a42f7a](https://github.com/kubedb/proxysql/commit/71a42f7a) Don't set annotation to AppBinding (#8)
- [282298cb](https://github.com/kubedb/proxysql/commit/282298cb) Fix UpsertDatabaseAnnotation() function (#4)
- [2ab9dddf](https://github.com/kubedb/proxysql/commit/2ab9dddf) Add license header to Makefiles (#6)
- [df135c08](https://github.com/kubedb/proxysql/commit/df135c08) Add install, uninstall and purge command in Makefile (#3)
- [73d3a845](https://github.com/kubedb/proxysql/commit/73d3a845) Update .gitignore
- [59a4e754](https://github.com/kubedb/proxysql/commit/59a4e754) Add Makefile (#2)
- [f3551ddc](https://github.com/kubedb/proxysql/commit/f3551ddc) Rename package path (#1)
- [56a241d6](https://github.com/kubedb/proxysql/commit/56a241d6) Use explicit IP whitelist instead of automatic IP whitelist (#151)
- [9f0b5ca3](https://github.com/kubedb/proxysql/commit/9f0b5ca3) Update to k8s 1.14.0 client libraries using go.mod (#147)
- [73ad7c30](https://github.com/kubedb/proxysql/commit/73ad7c30) Update changelog
- [ccc36b5c](https://github.com/kubedb/proxysql/commit/ccc36b5c) Update README.md
- [9769e8e1](https://github.com/kubedb/proxysql/commit/9769e8e1) Start next dev cycle
- [a3fa468a](https://github.com/kubedb/proxysql/commit/a3fa468a) Prepare release 0.5.0
- [6d8862de](https://github.com/kubedb/proxysql/commit/6d8862de) Mysql Group Replication tests (#146)
- [49544e55](https://github.com/kubedb/proxysql/commit/49544e55) Mysql Group Replication (#144)
- [a85d4b44](https://github.com/kubedb/proxysql/commit/a85d4b44) Revendor dependencies
- [9c538460](https://github.com/kubedb/proxysql/commit/9c538460) Changed Role to exclude psp without name (#143)
- [6cace93b](https://github.com/kubedb/proxysql/commit/6cace93b) Modify mutator validator names (#142)
- [da0c19b9](https://github.com/kubedb/proxysql/commit/da0c19b9) Update changelog
- [b79c80d6](https://github.com/kubedb/proxysql/commit/b79c80d6) Start next dev cycle
- [838d9459](https://github.com/kubedb/proxysql/commit/838d9459) Prepare release 0.4.0
- [bf0f2c14](https://github.com/kubedb/proxysql/commit/bf0f2c14) Added PSP names and init container image in testing framework (#141)
- [3d227570](https://github.com/kubedb/proxysql/commit/3d227570) Added PSP support for mySQL (#137)
- [7b766657](https://github.com/kubedb/proxysql/commit/7b766657) Don't inherit app.kubernetes.io labels from CRD into offshoots (#140)
- [29e23470](https://github.com/kubedb/proxysql/commit/29e23470) Support for init container (#139)
- [3e1556f6](https://github.com/kubedb/proxysql/commit/3e1556f6) Add role label to stats service (#138)
- [ee078af9](https://github.com/kubedb/proxysql/commit/ee078af9) Update changelog
- [978f1139](https://github.com/kubedb/proxysql/commit/978f1139) Update Kubernetes client libraries to 1.13.0 release (#136)
- [821f23d1](https://github.com/kubedb/proxysql/commit/821f23d1) Start next dev cycle
- [678b26aa](https://github.com/kubedb/proxysql/commit/678b26aa) Prepare release 0.3.0
- [40ad7a23](https://github.com/kubedb/proxysql/commit/40ad7a23) Initial RBAC support: create and use K8s service account for MySQL (#134)
- [98f03387](https://github.com/kubedb/proxysql/commit/98f03387) Revendor dependencies (#135)
- [dfe92615](https://github.com/kubedb/proxysql/commit/dfe92615) Revendor dependencies : Retry Failed Scheduler Snapshot (#133)
- [71f8a350](https://github.com/kubedb/proxysql/commit/71f8a350) Added ephemeral StorageType support (#132)
- [0a6b6e46](https://github.com/kubedb/proxysql/commit/0a6b6e46) Added support of MySQL 8.0.14 (#131)
- [99e57a9e](https://github.com/kubedb/proxysql/commit/99e57a9e) Use PVC spec from snapshot if provided (#130)
- [61497be6](https://github.com/kubedb/proxysql/commit/61497be6) Revendored and updated tests for 'Prevent prefix matching of multiple snapshots' (#129)
- [7eafe088](https://github.com/kubedb/proxysql/commit/7eafe088) Add certificate health checker (#128)
- [973ec416](https://github.com/kubedb/proxysql/commit/973ec416) Update E2E test: Env update is not restricted anymore (#127)
- [339975ff](https://github.com/kubedb/proxysql/commit/339975ff) Fix AppBinding (#126)
- [62050a72](https://github.com/kubedb/proxysql/commit/62050a72) Update changelog
- [2d454043](https://github.com/kubedb/proxysql/commit/2d454043) Prepare release 0.2.0
- [6941ea59](https://github.com/kubedb/proxysql/commit/6941ea59) Reuse event recorder (#125)
- [b77e66c4](https://github.com/kubedb/proxysql/commit/b77e66c4) OSM binary upgraded in mysql-tools (#123)
- [c9228086](https://github.com/kubedb/proxysql/commit/c9228086) Revendor dependencies (#124)
- [97837120](https://github.com/kubedb/proxysql/commit/97837120) Test for faulty snapshot (#122)
- [c3e995b6](https://github.com/kubedb/proxysql/commit/c3e995b6) Start next dev cycle
- [8a4f3b13](https://github.com/kubedb/proxysql/commit/8a4f3b13) Prepare release 0.2.0-rc.2
- [79942191](https://github.com/kubedb/proxysql/commit/79942191) Upgrade database secret keys (#121)
- [1747fdf5](https://github.com/kubedb/proxysql/commit/1747fdf5) Ignore mutation of fields to default values during update (#120)
- [d902d588](https://github.com/kubedb/proxysql/commit/d902d588) Support configuration options for exporter sidecar (#119)
- [dd7c3f44](https://github.com/kubedb/proxysql/commit/dd7c3f44) Use flags.DumpAll (#118)
- [bc1ef05b](https://github.com/kubedb/proxysql/commit/bc1ef05b) Start next dev cycle
- [9d33c1a0](https://github.com/kubedb/proxysql/commit/9d33c1a0) Prepare release 0.2.0-rc.1
- [b076e141](https://github.com/kubedb/proxysql/commit/b076e141) Apply cleanup (#117)
- [7dc5641f](https://github.com/kubedb/proxysql/commit/7dc5641f) Set periodic analytics (#116)
- [90ea6acc](https://github.com/kubedb/proxysql/commit/90ea6acc) Introduce AppBinding support (#115)
- [a882d76a](https://github.com/kubedb/proxysql/commit/a882d76a) Fix Analytics (#114)
- [0961009c](https://github.com/kubedb/proxysql/commit/0961009c) Error out from cron job for deprecated dbversion (#113)
- [da1f4e27](https://github.com/kubedb/proxysql/commit/da1f4e27) Add CRDs without observation when operator starts (#112)
- [0a754d2f](https://github.com/kubedb/proxysql/commit/0a754d2f) Update changelog
- [b09bc6e1](https://github.com/kubedb/proxysql/commit/b09bc6e1) Start next dev cycle
- [0d467ccb](https://github.com/kubedb/proxysql/commit/0d467ccb) Prepare release 0.2.0-rc.0
- [c757007a](https://github.com/kubedb/proxysql/commit/c757007a) Merge commit 'cc6607a3589a79a5e61bb198d370ea0ae30b9d09'
- [ddfe4be1](https://github.com/kubedb/proxysql/commit/ddfe4be1) Support custom user passowrd for backup (#111)
- [8c84ba20](https://github.com/kubedb/proxysql/commit/8c84ba20) Support providing resources for monitoring container (#110)
- [7bcfbc48](https://github.com/kubedb/proxysql/commit/7bcfbc48) Update kubernetes client libraries to 1.12.0 (#109)
- [145bba2b](https://github.com/kubedb/proxysql/commit/145bba2b) Add validation webhook xray (#108)
- [6da1887f](https://github.com/kubedb/proxysql/commit/6da1887f) Various Fixes (#107)
- [111519e9](https://github.com/kubedb/proxysql/commit/111519e9) Merge ports from service template (#105)
- [38147ef1](https://github.com/kubedb/proxysql/commit/38147ef1) Replace doNotPause with TerminationPolicy = DoNotTerminate (#104)
- [e28ebc47](https://github.com/kubedb/proxysql/commit/e28ebc47) Pass resources to NamespaceValidator (#103)
- [aed12bf5](https://github.com/kubedb/proxysql/commit/aed12bf5) Various fixes (#102)
- [3d372ef6](https://github.com/kubedb/proxysql/commit/3d372ef6) Support Livecycle hook and container probes (#101)
- [b6ef6887](https://github.com/kubedb/proxysql/commit/b6ef6887) Check if Kubernetes version is supported before running operator (#100)
- [d89e7783](https://github.com/kubedb/proxysql/commit/d89e7783) Update package alias (#99)
- [f0b44b3a](https://github.com/kubedb/proxysql/commit/f0b44b3a) Start next dev cycle
- [a79ff03b](https://github.com/kubedb/proxysql/commit/a79ff03b) Prepare release 0.2.0-beta.1
- [0d8d3cca](https://github.com/kubedb/proxysql/commit/0d8d3cca) Revendor api (#98)
- [2f850243](https://github.com/kubedb/proxysql/commit/2f850243) Fix tests (#97)
- [4ced0bfe](https://github.com/kubedb/proxysql/commit/4ced0bfe) Revendor api for catalog apigroup (#96)
- [e7695400](https://github.com/kubedb/proxysql/commit/e7695400) Update chanelog
- [8e358aea](https://github.com/kubedb/proxysql/commit/8e358aea) Use --pull flag with docker build (#20) (#95)
- [d2a97d90](https://github.com/kubedb/proxysql/commit/d2a97d90) Merge commit '16c769ee4686576f172a6b79a10d25bfd79ca4a4'
- [d1fe8a8a](https://github.com/kubedb/proxysql/commit/d1fe8a8a) Start next dev cycle
- [04eb9bb5](https://github.com/kubedb/proxysql/commit/04eb9bb5) Prepare release 0.2.0-beta.0
- [9dfea960](https://github.com/kubedb/proxysql/commit/9dfea960) Pass extra args to tools.sh (#93)
- [47dd3cad](https://github.com/kubedb/proxysql/commit/47dd3cad) Don't try to wipe out Snapshot data for Local backend (#92)
- [9c4d485b](https://github.com/kubedb/proxysql/commit/9c4d485b) Add missing alt-tag docker folder mysql-tools images (#91)
- [be72f784](https://github.com/kubedb/proxysql/commit/be72f784) Use suffix for updated DBImage & Stop working for deprecated *Versions (#90)
- [05c8f14d](https://github.com/kubedb/proxysql/commit/05c8f14d) Search used secrets within same namespace of DB object (#89)
- [0d94c946](https://github.com/kubedb/proxysql/commit/0d94c946) Support Termination Policy (#88)
- [8775ddf7](https://github.com/kubedb/proxysql/commit/8775ddf7) Update builddeps.sh
- [796c93da](https://github.com/kubedb/proxysql/commit/796c93da) Revendor k8s.io/apiserver (#87)
- [5a1e3f57](https://github.com/kubedb/proxysql/commit/5a1e3f57) Revendor kubernetes-1.11.3 (#86)
- [809a3c49](https://github.com/kubedb/proxysql/commit/809a3c49) Support UpdateStrategy (#84)
- [372c52ef](https://github.com/kubedb/proxysql/commit/372c52ef) Add TerminationPolicy for databases (#83)
- [c01b55e8](https://github.com/kubedb/proxysql/commit/c01b55e8) Revendor api (#82)
- [5e196b95](https://github.com/kubedb/proxysql/commit/5e196b95) Use IntHash as status.observedGeneration (#81)
- [2da3bb1b](https://github.com/kubedb/proxysql/commit/2da3bb1b) fix github status (#80)
- [121d0a98](https://github.com/kubedb/proxysql/commit/121d0a98) Update pipeline (#79)
- [532e3137](https://github.com/kubedb/proxysql/commit/532e3137) Fix E2E test for minikube (#78)
- [0f107815](https://github.com/kubedb/proxysql/commit/0f107815) Update pipeline (#77)
- [851679e2](https://github.com/kubedb/proxysql/commit/851679e2) Migrate MySQL (#75)
- [0b997855](https://github.com/kubedb/proxysql/commit/0b997855) Use official exporter image (#74)
- [702d5736](https://github.com/kubedb/proxysql/commit/702d5736) Fix uninstall for concourse (#70)
- [9ee88bd2](https://github.com/kubedb/proxysql/commit/9ee88bd2) Update status.ObservedGeneration for failure phase (#73)
- [559cdb6a](https://github.com/kubedb/proxysql/commit/559cdb6a) Keep track of ObservedGenerationHash (#72)
- [61c8b898](https://github.com/kubedb/proxysql/commit/61c8b898) Use NewObservableHandler (#71)
- [421274dc](https://github.com/kubedb/proxysql/commit/421274dc) Merge commit '887037c7e36289e3135dda99346fccc7e2ce303b'
- [6a41d9bc](https://github.com/kubedb/proxysql/commit/6a41d9bc) Fix uninstall for concourse (#69)
- [f1af09db](https://github.com/kubedb/proxysql/commit/f1af09db) Update README.md
- [bf3f1823](https://github.com/kubedb/proxysql/commit/bf3f1823) Revise immutable spec fields (#68)
- [26adec3b](https://github.com/kubedb/proxysql/commit/26adec3b) Merge commit '5f83049fc01dc1d0709ac0014d6f3a0f74a39417'
- [31a97820](https://github.com/kubedb/proxysql/commit/31a97820) Support passing args via PodTemplate (#67)
- [60f4ee23](https://github.com/kubedb/proxysql/commit/60f4ee23) Introduce storageType : ephemeral (#66)
- [bfd3fcd6](https://github.com/kubedb/proxysql/commit/bfd3fcd6) Add support for running tests on cncf cluster (#63)
- [fba47b19](https://github.com/kubedb/proxysql/commit/fba47b19) Merge commit 'e010cbb302c8d59d4cf69dd77085b046ff423b78'
- [6be96ce0](https://github.com/kubedb/proxysql/commit/6be96ce0) Revendor api (#65)
- [0f629ab3](https://github.com/kubedb/proxysql/commit/0f629ab3) Keep track of observedGeneration in status (#64)
- [c9a9596f](https://github.com/kubedb/proxysql/commit/c9a9596f) Separate StatsService for monitoring (#62)
- [62854641](https://github.com/kubedb/proxysql/commit/62854641) Use MySQLVersion for MySQL images (#61)
- [3c170c56](https://github.com/kubedb/proxysql/commit/3c170c56) Use updated crd spec (#60)
- [873c285e](https://github.com/kubedb/proxysql/commit/873c285e) Rename OffshootLabels to OffshootSelectors (#59)
- [2fd02169](https://github.com/kubedb/proxysql/commit/2fd02169) Revendor api (#58)
- [a127d6cd](https://github.com/kubedb/proxysql/commit/a127d6cd) Use kmodules monitoring and objectstore api (#57)
- [2f79a038](https://github.com/kubedb/proxysql/commit/2f79a038) Support custom configuration (#52)
- [49c67f00](https://github.com/kubedb/proxysql/commit/49c67f00) Merge commit '44e6d4985d93556e39ddcc4677ada5437fc5be64'
- [fb28bc6c](https://github.com/kubedb/proxysql/commit/fb28bc6c) Refactor concourse scripts (#56)
- [4de4ced1](https://github.com/kubedb/proxysql/commit/4de4ced1) Fix command `./hack/make.py test e2e` (#55)
- [3082123e](https://github.com/kubedb/proxysql/commit/3082123e) Set generated binary name to my-operator (#54)
- [5698f314](https://github.com/kubedb/proxysql/commit/5698f314) Don't add admission/v1beta1 group as a prioritized version (#53)
- [696135d5](https://github.com/kubedb/proxysql/commit/696135d5) Fix travis build (#48)
- [c519ef89](https://github.com/kubedb/proxysql/commit/c519ef89) Format shell script (#51)
- [c93e2f40](https://github.com/kubedb/proxysql/commit/c93e2f40) Enable status subresource for crds (#50)
- [edd951ca](https://github.com/kubedb/proxysql/commit/edd951ca) Update client-go to v8.0.0 (#49)
- [520597a6](https://github.com/kubedb/proxysql/commit/520597a6) Merge commit '71850e2c90cda8fc588b7dedb340edf3d316baea'
- [f1549e95](https://github.com/kubedb/proxysql/commit/f1549e95) Support ENV variables in CRDs (#46)
- [67f37780](https://github.com/kubedb/proxysql/commit/67f37780) Updated osm version to 0.7.1 (#47)
- [10e309c0](https://github.com/kubedb/proxysql/commit/10e309c0) Prepare release 0.1.0
- [62a8fbbd](https://github.com/kubedb/proxysql/commit/62a8fbbd) Fixed missing error return (#45)
- [8c05bb83](https://github.com/kubedb/proxysql/commit/8c05bb83) Revendor dependencies (#44)
- [ca811a2e](https://github.com/kubedb/proxysql/commit/ca811a2e) Fix release script (#43)
- [b79541f6](https://github.com/kubedb/proxysql/commit/b79541f6) Add changelog (#42)
- [a2d13c82](https://github.com/kubedb/proxysql/commit/a2d13c82) Concourse (#41)
- [95b2186e](https://github.com/kubedb/proxysql/commit/95b2186e) Fixed kubeconfig plugin for Cloud Providers && Storage is required for MySQL (#40)
- [37762093](https://github.com/kubedb/proxysql/commit/37762093) Refactored E2E testing to support E2E testing with admission webhook in cloud (#38)
- [b6fe72ca](https://github.com/kubedb/proxysql/commit/b6fe72ca) Remove lost+found directory before initializing mysql (#39)
- [18ebb959](https://github.com/kubedb/proxysql/commit/18ebb959) Skip delete requests for empty resources (#37)
- [eeb7add0](https://github.com/kubedb/proxysql/commit/eeb7add0) Don't panic if admission options is nil (#36)
- [ccb59db0](https://github.com/kubedb/proxysql/commit/ccb59db0) Disable admission controllers for webhook server (#35)
- [b1c6c149](https://github.com/kubedb/proxysql/commit/b1c6c149) Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 (#34)
- [b1890f7c](https://github.com/kubedb/proxysql/commit/b1890f7c) Update client-go to 7.0.0 (#33)
- [08c81726](https://github.com/kubedb/proxysql/commit/08c81726) Added update script for mysql-tools:8 (#32)
- [4bbe6c9f](https://github.com/kubedb/proxysql/commit/4bbe6c9f) Added support of mysql:5.7 (#31)
- [e657f512](https://github.com/kubedb/proxysql/commit/e657f512) Add support for one informer and N-eventHandler for snapshot, dromantDB and Job (#30)
- [bbcd48d6](https://github.com/kubedb/proxysql/commit/bbcd48d6) Use metrics from kube apiserver (#29)
- [1687e197](https://github.com/kubedb/proxysql/commit/1687e197) Bundle webhook server and Use SharedInformerFactory (#28)
- [cd0efc00](https://github.com/kubedb/proxysql/commit/cd0efc00) Move MySQL AdmissionWebhook packages into MySQL repository (#27)
- [46065e18](https://github.com/kubedb/proxysql/commit/46065e18) Use mysql:8.0.3 image as mysql:8.0 (#26)
- [1b73529f](https://github.com/kubedb/proxysql/commit/1b73529f) Update README.md
- [62eaa397](https://github.com/kubedb/proxysql/commit/62eaa397) Update README.md
- [c53704c7](https://github.com/kubedb/proxysql/commit/c53704c7) Remove Docker pull count
- [b9ec877e](https://github.com/kubedb/proxysql/commit/b9ec877e) Add travis yaml (#25)
- [ade3571c](https://github.com/kubedb/proxysql/commit/ade3571c) Start next dev cycle
- [b4b749df](https://github.com/kubedb/proxysql/commit/b4b749df) Prepare release 0.1.0-beta.2
- [4d46d95d](https://github.com/kubedb/proxysql/commit/4d46d95d) Migrating to apps/v1 (#23)
- [5ee1ac8c](https://github.com/kubedb/proxysql/commit/5ee1ac8c) Update validation (#22)
- [dd023c50](https://github.com/kubedb/proxysql/commit/dd023c50)  Fix dormantDB matching: pass same type to Equal method (#21)
- [37a1e4fd](https://github.com/kubedb/proxysql/commit/37a1e4fd) Use official code generator scripts (#20)
- [485d3d7c](https://github.com/kubedb/proxysql/commit/485d3d7c) Fixed dormantdb matching & Raised throttling time & Fixed MySQL version Checking (#19)
- [6db2ae8d](https://github.com/kubedb/proxysql/commit/6db2ae8d) Prepare release 0.1.0-beta.1
- [ebbfec2f](https://github.com/kubedb/proxysql/commit/ebbfec2f) converted to k8s 1.9 & Improved InitSpec in DormantDB & Added support for Job watcher & Improved Tests (#17)
- [a484e0e5](https://github.com/kubedb/proxysql/commit/a484e0e5) Fixed logger, analytics and removed rbac stuff (#16)
- [7aa2d1d2](https://github.com/kubedb/proxysql/commit/7aa2d1d2) Add rbac stuffs for mysql-exporter (#15)
- [078098c8](https://github.com/kubedb/proxysql/commit/078098c8)  Review Mysql docker images and Fixed monitring (#14)
- [6877108a](https://github.com/kubedb/proxysql/commit/6877108a) Update README.md
- [1f84a5da](https://github.com/kubedb/proxysql/commit/1f84a5da) Start next dev cycle
- [2f1e4b7d](https://github.com/kubedb/proxysql/commit/2f1e4b7d) Prepare release 0.1.0-beta.0
- [dce1e88e](https://github.com/kubedb/proxysql/commit/dce1e88e) Add release script
- [60ed55cb](https://github.com/kubedb/proxysql/commit/60ed55cb) Rename ms-operator to my-operator (#13)
- [5451d166](https://github.com/kubedb/proxysql/commit/5451d166) Fix Analytics and pass client-id as ENV to Snapshot Job (#12)
- [788ae178](https://github.com/kubedb/proxysql/commit/788ae178) update docker image validation (#11)
- [c966efd5](https://github.com/kubedb/proxysql/commit/c966efd5) Add docker-registry and WorkQueue  (#10)
- [be340103](https://github.com/kubedb/proxysql/commit/be340103) Set client id for analytics (#9)
- [ca11f683](https://github.com/kubedb/proxysql/commit/ca11f683) Fix CRD Registration (#8)
- [2f95c13d](https://github.com/kubedb/proxysql/commit/2f95c13d) Update issue repo link
- [6fffa713](https://github.com/kubedb/proxysql/commit/6fffa713) Update pkg paths to kubedb org (#7)
- [2d4d5c44](https://github.com/kubedb/proxysql/commit/2d4d5c44) Assign default Prometheus Monitoring Port (#6)
- [a7595613](https://github.com/kubedb/proxysql/commit/a7595613) Add Snapshot Backup, Restore and Backup-Scheduler (#4)
- [17a782c6](https://github.com/kubedb/proxysql/commit/17a782c6) Update Dockerfile
- [e92bfec9](https://github.com/kubedb/proxysql/commit/e92bfec9) Add mysql-util docker image (#5)
- [2a4b25ac](https://github.com/kubedb/proxysql/commit/2a4b25ac) Mysql db - Inititalizing  (#2)
- [cbfbc878](https://github.com/kubedb/proxysql/commit/cbfbc878) Update README.md
- [01cab651](https://github.com/kubedb/proxysql/commit/01cab651) Update README.md
- [0aa81cdf](https://github.com/kubedb/proxysql/commit/0aa81cdf) Use client-go 5.x
- [3de10d7f](https://github.com/kubedb/proxysql/commit/3de10d7f) Update ./hack folder (#3)
- [46f05b1f](https://github.com/kubedb/proxysql/commit/46f05b1f) Add skeleton for mysql (#1)
- [73147dba](https://github.com/kubedb/proxysql/commit/73147dba) Merge commit 'be70502b4993171bbad79d2ff89a9844f1c24caa' as 'hack/libbuild'



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.7.0-beta.1](https://github.com/kubedb/redis/releases/tag/v0.7.0-beta.1)

- [768962f4](https://github.com/kubedb/redis/commit/768962f4) Prepare for release v0.7.0-beta.1 (#173)
- [9efbb8e4](https://github.com/kubedb/redis/commit/9efbb8e4) include Makefile.env (#171)
- [b343c559](https://github.com/kubedb/redis/commit/b343c559) Update License (#170)
- [d666ac18](https://github.com/kubedb/redis/commit/d666ac18) Update to Kubernetes v1.18.3 (#169)
- [602354f6](https://github.com/kubedb/redis/commit/602354f6) Update ci.yml
- [59f2d238](https://github.com/kubedb/redis/commit/59f2d238) Update update-release-tracker.sh
- [64c96db5](https://github.com/kubedb/redis/commit/64c96db5) Update update-release-tracker.sh
- [49cd15a9](https://github.com/kubedb/redis/commit/49cd15a9) Add script to update release tracker on pr merge (#167)
- [c711be8f](https://github.com/kubedb/redis/commit/c711be8f) chore: replica alert typo (#166)
- [2d752316](https://github.com/kubedb/redis/commit/2d752316) Update .kodiak.toml
- [ea3b206d](https://github.com/kubedb/redis/commit/ea3b206d) Various fixes (#165)
- [e441809c](https://github.com/kubedb/redis/commit/e441809c) Update to Kubernetes v1.18.3 (#164)
- [1e5ecfb7](https://github.com/kubedb/redis/commit/1e5ecfb7) Update to Kubernetes v1.18.3
- [742679dd](https://github.com/kubedb/redis/commit/742679dd) Create .kodiak.toml
- [2eb77b80](https://github.com/kubedb/redis/commit/2eb77b80) Update apis (#163)
- [7cf9e7d3](https://github.com/kubedb/redis/commit/7cf9e7d3) Use CRD v1 for Kubernetes >= 1.16 (#162)
- [bf072134](https://github.com/kubedb/redis/commit/bf072134) Update kind command
- [cb2a748d](https://github.com/kubedb/redis/commit/cb2a748d) Update dependencies
- [a30cd6eb](https://github.com/kubedb/redis/commit/a30cd6eb) Update to Kubernetes v1.18.3 (#161)
- [9cdac95f](https://github.com/kubedb/redis/commit/9cdac95f) Fix e2e tests (#160)
- [429141b4](https://github.com/kubedb/redis/commit/429141b4) Revendor kubedb.dev/apimachinery@master (#159)
- [664c086b](https://github.com/kubedb/redis/commit/664c086b) Use recommended kubernetes app labels
- [2e6a2f03](https://github.com/kubedb/redis/commit/2e6a2f03) Update crazy-max/ghaction-docker-buildx flag
- [88417e86](https://github.com/kubedb/redis/commit/88417e86) Pass annotations from CRD to AppBinding (#158)
- [84167d7a](https://github.com/kubedb/redis/commit/84167d7a) Trigger the workflow on push or pull request
- [2f43dd9a](https://github.com/kubedb/redis/commit/2f43dd9a) Use helm --wait
- [36399173](https://github.com/kubedb/redis/commit/36399173) Use updated operator labels in e2e tests (#156)
- [c6582491](https://github.com/kubedb/redis/commit/c6582491) Update CHANGELOG.md
- [197b4973](https://github.com/kubedb/redis/commit/197b4973) Support PodAffinity Templating (#155)
- [cdfbb77d](https://github.com/kubedb/redis/commit/cdfbb77d) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#154)
- [c1db4c43](https://github.com/kubedb/redis/commit/c1db4c43) Version update to resolve security issue in github.com/apache/th… (#153)
- [7acc502b](https://github.com/kubedb/redis/commit/7acc502b) Use rancher/local-path-provisioner@v0.0.12 (#152)
- [d00f765e](https://github.com/kubedb/redis/commit/d00f765e) Introduce spec.halted and removed dormant crd (#151)
- [9ed1d97e](https://github.com/kubedb/redis/commit/9ed1d97e) Add `Pause` Feature (#150)
- [39ed60c4](https://github.com/kubedb/redis/commit/39ed60c4) Refactor CI pipeline to build once (#149)
- [1707e0c7](https://github.com/kubedb/redis/commit/1707e0c7) Update kubernetes client-go to 1.16.3 (#148)
- [dcbb4be4](https://github.com/kubedb/redis/commit/dcbb4be4) Update catalog values for make install command
- [9fa3ef1c](https://github.com/kubedb/redis/commit/9fa3ef1c) Update catalog values for make install command (#147)
- [44538409](https://github.com/kubedb/redis/commit/44538409) Use charts to install operator (#146)
- [05e3b95a](https://github.com/kubedb/redis/commit/05e3b95a) Matrix test for github actions (#145)
- [e76f96f6](https://github.com/kubedb/redis/commit/e76f96f6) Add add-license make target
- [6ccd651c](https://github.com/kubedb/redis/commit/6ccd651c) Update Makefile
- [2a56f27f](https://github.com/kubedb/redis/commit/2a56f27f) Add license header to files (#144)
- [5ce5e5e0](https://github.com/kubedb/redis/commit/5ce5e5e0) Run e2e tests in parallel (#142)
- [77012ddf](https://github.com/kubedb/redis/commit/77012ddf) Use log.Fatal instead of Must() (#143)
- [aa7f1673](https://github.com/kubedb/redis/commit/aa7f1673) Enable make ci (#141)
- [abd6a605](https://github.com/kubedb/redis/commit/abd6a605) Remove EnableStatusSubresource (#140)
- [08cfe0ca](https://github.com/kubedb/redis/commit/08cfe0ca) Fix tests for github actions (#139)
- [09e72f63](https://github.com/kubedb/redis/commit/09e72f63) Prepend redis.conf to args list (#136)
- [101afa35](https://github.com/kubedb/redis/commit/101afa35) Run e2e tests using GitHub actions (#137)
- [bbf5cb9f](https://github.com/kubedb/redis/commit/bbf5cb9f) Validate DBVersionSpecs and fixed broken build (#138)
- [26f0c88b](https://github.com/kubedb/redis/commit/26f0c88b) Update go.yml
- [9dab8c06](https://github.com/kubedb/redis/commit/9dab8c06) Enable GitHub actions
- [6a722f20](https://github.com/kubedb/redis/commit/6a722f20) Update changelog




