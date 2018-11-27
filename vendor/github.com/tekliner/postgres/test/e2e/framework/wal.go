package framework

import (
	"time"

	"github.com/graymeta/stow"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/postgres/pkg/controller"
	. "github.com/onsi/gomega"
	storage "kmodules.xyz/objectstore-api/osm"
)

func (f *Framework) EventuallyWalDataFound(postgres *api.Postgres) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			found, err := f.checkWalData(postgres)
			Expect(err).NotTo(HaveOccurred())
			return found
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) checkWalData(postgres *api.Postgres) (bool, error) {
	cfg, err := storage.NewOSMContext(f.kubeClient, *postgres.Spec.Archiver.Storage, postgres.Namespace)
	if err != nil {
		return false, err
	}

	loc, err := stow.Dial(cfg.Provider, cfg.Config)
	if err != nil {
		return false, err
	}
	containerID, err := postgres.Spec.Archiver.Storage.Container()
	if err != nil {
		return false, err
	}
	container, err := loc.Container(containerID)
	if err != nil {
		return false, err
	}

	prefix := controller.WalDataDir(postgres)
	cursor := stow.CursorStart
	totalItem := 0
	for {
		items, next, err := container.Items(prefix, cursor, 50)
		if err != nil {
			return false, err
		}

		totalItem = totalItem + len(items)

		cursor = next
		if stow.IsCursorEnd(cursor) {
			break
		}
	}

	return totalItem != 0, nil
}
