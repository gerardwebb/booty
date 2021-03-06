package components

import (
	"go.amplifyedge.org/booty-v2/dep"
	"go.amplifyedge.org/booty-v2/internal/downloader"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/store"
	"go.amplifyedge.org/booty-v2/internal/update"
	"os"
	"path/filepath"
)

const (
	jbBaseRepo = "github.com/jsonnet-bundler/jsonnet-bundler"
	jbRepo     = "https://" + jbBaseRepo
)

type Jb struct {
	version update.Version
	db      store.Storer
}

func (j *Jb) IsService() bool {
	return false
}

func (j *Jb) IsDev() bool {
	return true
}

func NewJb(db store.Storer) *Jb {
	return &Jb{
		db: db,
	}
}

func (j *Jb) SetVersion(v update.Version) {
	j.version = v
}

func (j *Jb) Name() string {
	return "jb"
}

func (j *Jb) Version() update.Version {
	if v := commonGetVersion(j, j.db); v != nil {
		return *v
	}
	return j.version
}

func (j *Jb) Download() error {
	targetDir := getDlPath(j.Name(), j.version.String())
	return downloader.Download(jbBaseRepo+"?ref=v"+j.version.String(), targetDir)
}

func (j *Jb) Dependencies() []dep.Component {
	return nil
}

func (j *Jb) Install() error {
	targetDir := getDlPath(j.Name(), j.version.String())
	var err error
	binDir := osutil.GetBinDir()
	if err = os.Chdir(targetDir); err != nil {
		return err
	}
	if err = osutil.Exec("make", "static"); err != nil {
		return err
	}

	filesMap := map[string][]interface{}{
		filepath.Join(targetDir, "_output", j.Name()): {filepath.Join(binDir, j.Name()), 0755},
	}

	// copy file to the bin directory
	ip, err := commonInstall(j, filesMap)
	if err != nil {
		return err
	}

	if err = j.db.New(ip); err != nil {
		return err
	}
	return nil
}

func (j *Jb) Uninstall() error {
	var err error
	var pkg *store.InstalledPackage
	pkg, err = j.db.Get(j.Name())
	if err != nil {
		return err
	}
	var filesList []string
	for k := range pkg.FilesMap {
		filesList = append(filesList, k)
	}
	for _, file := range filesList {
		if err = os.RemoveAll(file); err != nil {
			return err
		}
	}
	if err = j.db.Delete(j.Name()); err != nil {
		return err
	}
	return os.RemoveAll(getDlPath(j.Name(), j.version.String()))
}

func (j *Jb) Run(args ...string) error {
	return osutil.Exec(filepath.Join(osutil.GetBinDir(), j.Name()), args...)
}

func (j *Jb) Update(version update.Version) error {
	return commonUpdate(j, version)
}

func (j *Jb) RunStop() error {
	return nil
}

func (j *Jb) Backup() error {
	return nil
}

func (j *Jb) RepoUrl() update.RepositoryURL {
	return jbRepo
}
