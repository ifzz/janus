package oauth

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"net/url"

	log "github.com/Sirupsen/logrus"
)

// FileSystemRepository represents a mongodb repository
type FileSystemRepository struct {
	sync.Mutex
	servers map[string]*OAuth
}

// NewFileSystemRepository creates a mongo OAuth Server repo
func NewFileSystemRepository(dir string) (*FileSystemRepository, error) {
	repo := &FileSystemRepository{servers: make(map[string]*OAuth)}
	// Grab json files from directory
	files, err := ioutil.ReadDir(dir)
	if nil != err {
		return nil, err
	}

	for _, f := range files {
		if strings.Contains(f.Name(), ".json") {
			filePath := filepath.Join(dir, f.Name())
			log.WithField("path", filePath).Info("Loading OAuth Server definition from file")
			appConfigBody, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.WithError(err).WithField("path", filePath).Error("Couldn't load oauth server definition file")
				return nil, err
			}

			definition := repo.parseDefinition(appConfigBody)
			err = repo.Add(definition)
			if err != nil {
				log.WithError(err).Error("Can't add the oauth server to the repository")
				return nil, err
			}
		}
	}

	return repo, nil
}

// FindAll fetches all the OAuth Servers available
func (r *FileSystemRepository) FindAll() ([]*OAuth, error) {
	var servers []*OAuth
	for _, server := range r.servers {
		servers = append(servers, server)
	}

	return servers, nil
}

// FindByName find an OAuth Server by name
func (r *FileSystemRepository) FindByName(name string) (*OAuth, error) {
	server, ok := r.servers[name]
	if false == ok {
		return nil, ErrOauthServerNotFound
	}

	return server, nil
}

// Add adds an OAuth Server to the repository
func (r *FileSystemRepository) Add(server *OAuth) error {
	r.Lock()
	defer r.Unlock()

	r.servers[server.Name] = server

	return nil
}

// Remove removes an OAuth Server from the repository
func (r *FileSystemRepository) Remove(name string) error {
	r.Lock()
	defer r.Unlock()

	delete(r.servers, name)
	return nil
}

// FindByTokenURL returns OAuth Server records with corresponding token url
func (r *FileSystemRepository) FindByTokenURL(url url.URL) (*OAuth, error) {
	for _, server := range r.servers {
		if server.Endpoints.Token.UpstreamURL == url.String() {
			return server, nil
		}
	}

	return nil, ErrOauthServerNotFound
}

func (r *FileSystemRepository) parseDefinition(apiDef []byte) *OAuth {
	appConfig := &OAuth{}
	if err := json.Unmarshal(apiDef, appConfig); err != nil {
		log.Error("[RPC] --> Couldn't unmarshal api configuration: ", err)
	}

	return appConfig
}
