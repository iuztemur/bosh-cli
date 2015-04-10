package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	. "github.com/cloudfoundry/bosh-init/config"
	"github.com/cloudfoundry/bosh-init/release"
	fakerelease "github.com/cloudfoundry/bosh-init/release/fakes"
)

var _ = Describe("ReleaseRepo", rootDesc)

func rootDesc() {
	var (
		repo              ReleaseRepo
		configService     DeploymentConfigService
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		fakeUUIDGenerator.GeneratedUUID = "fake-uuid"
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		configService.Load()
		repo = NewReleaseRepo(configService, fakeUUIDGenerator)
	})

	Describe("List", func() {
		Context("when a current release exists", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "fake-guid-a", Name: "fake-name-a", Version: "fake-version-a"},
					ReleaseRecord{ID: "fake-guid-b", Name: "fake-name-b", Version: "fake-version-b"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns existing release", func() {
				records, err := repo.List()
				Expect(err).ToNot(HaveOccurred())
				Expect(records).To(Equal([]ReleaseRecord{
					{
						ID:      "fake-guid-a",
						Name:    "fake-name-a",
						Version: "fake-version-a",
					},
					{
						ID:      "fake-guid-b",
						Name:    "fake-name-b",
						Version: "fake-version-b",
					},
				}))
			})
		})

		Context("when there are no releases recorded", func() {
			It("returns not found", func() {
				records, err := repo.List()
				Expect(err).ToNot(HaveOccurred())
				Expect(records).To(HaveLen(0))
			})
		})

		Context("when the config service fails to load", func() {
			BeforeEach(func() {
				fs.ReadFileError = errors.New("kaboom")
			})

			It("returns an error", func() {
				_, err := repo.List()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Loading existing config"))
			})
		})
	})

	Describe("Update", func() {
		Context("when there are no existing releases", func() {
			It("saves the provided releases to the config file", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).ToNot(HaveOccurred())
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.Releases).To(ConsistOf(
					ReleaseRecord{ID: "fake-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "fake-uuid", Name: "name2", Version: "2"},
				))
			})
		})

		Context("when the existing releases exactly match the provided releases", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "old-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "old-uuid", Name: "name2", Version: "2"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when existing versions differ from the provided release versions", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "old-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "old-uuid", Name: "name2", Version: "3"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the provided releases to the config file", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).ToNot(HaveOccurred())
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.Releases).To(ConsistOf(
					ReleaseRecord{ID: "fake-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "fake-uuid", Name: "name2", Version: "2"},
				))
			})
		})

		Context("when existing names differ from the provided release names", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "old-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "old-uuid", Name: "other-name", Version: "2"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the provided releases to the config file", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).ToNot(HaveOccurred())
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.Releases).To(ConsistOf(
					ReleaseRecord{ID: "fake-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "fake-uuid", Name: "name2", Version: "2"},
				))
			})
		})

		Context("when a release is removed", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "old-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "old-uuid", Name: "name2", Version: "2"},
					ReleaseRecord{ID: "old-uuid", Name: "name3", Version: "3"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the provided releases to the config file", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).ToNot(HaveOccurred())
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.Releases).To(ConsistOf(
					ReleaseRecord{ID: "fake-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "fake-uuid", Name: "name2", Version: "2"},
				))
			})
		})

		Context("when a release is added", func() {
			BeforeEach(func() {
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				conf.Releases = []ReleaseRecord{
					ReleaseRecord{ID: "old-uuid", Name: "name1", Version: "1"},
				}
				err = configService.Save(conf)
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the provided releases to the config file", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).ToNot(HaveOccurred())
				conf, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(conf.Releases).To(ConsistOf(
					ReleaseRecord{ID: "fake-uuid", Name: "name1", Version: "1"},
					ReleaseRecord{ID: "fake-uuid", Name: "name2", Version: "2"},
				))
			})
		})

		Context("when the config service fails to save", func() {
			BeforeEach(func() {
				fs.WriteFileError = errors.New("kaboom")
			})

			It("returns an error", func() {
				err := repo.Update([]release.Release{
					fakerelease.New("name1", "1"),
					fakerelease.New("name2", "2"),
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
}
