package zim

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tim-st/go-zim"
)

var (
	Log *logrus.Logger
)

// Handler manage zim files
type Handler struct {
	rawFiles []string
	files    []File

	// Cache for faster file search
	fileCache map[string]*File

	Mx sync.Mutex
}

// New create a new zimservice
func New(files []string) *Handler {
	return &Handler{
		rawFiles:  files,
		fileCache: make(map[string]*File),
	}
}

// Start starts the zimservice
func (zs *Handler) Start(indexPath string) error {

	// Load all zimfiles in given directorys
	if err := zs.loadFiles(); err != nil {
		return err
	}

	Log.Infof("%d zim file(s) found", len(zs.files))

	_, folderErr := os.Stat(indexPath)
	if errors.Is(folderErr, os.ErrNotExist) {
		folderErr := os.Mkdir(indexPath, os.ModePerm)
		if folderErr != nil {
			Log.Error(folderErr)
		}
	}

	// TODO add bleve indexing

	// Set to true for bleve indexing
	err := zs.GenerateIndex(indexPath, false)
	if err == nil {
		Log.Info("Indexing successful")
	}

	return err
}

// GetFiles in dir
func (zs *Handler) GetFiles() []File {
	return zs.files
}

// Load all files in given Dir
func (zs *Handler) loadFiles() error {
	var success, errs int

	for _, file := range zs.rawFiles {
		fileInfo, err := os.Stat(file)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
				// Ignore non regular files
				if !info.IsDir() && !strings.HasSuffix(path, ".ix") && !strings.HasSuffix(path, ".ix.db") {
					// We want to use the real
					// path ond disk
					realPath := path

					// Follow sysmlinks
					if info.Mode()&os.ModeSymlink == os.ModeSymlink {
						// Follow link
						path, err = os.Readlink(path)
						if err != nil {
							return err
						}
					}

					// Try to open any file
					f, err := zim.Open(path)
					if err != nil {
						errs++
						Log.Error(errors.Wrap(err, path))

						// Ignore errors for now
						return nil
					}

					zs.files = append(zs.files, File{
						File: f,
						Path: realPath,
					})
					success++
				}

				return nil
			})
		} else {
			f, err := zim.Open(file)
			if err != nil {
				errs++
				Log.Error(errors.Wrap(err, file))

				// Ignore errors for now
				return nil
			}

			zs.files = append(zs.files, File{
				File: f,
				Path: file,
			})
			success++
		}
	}

	if success == 0 && errs > 0 {
		Log.Fatal("Too many errors")
	}

	return nil
}

// FindWikiFile by ID. File gets cached into a map
func (zs *Handler) FindWikiFile(zimFileID string) *File {
	if fil, has := zs.fileCache[zimFileID]; has {
		return fil
	}

	// Loop all files and find matching
	for i := range zs.files {
		file := &zs.files[i]

		if file.GetID() == zimFileID {
			zs.fileCache[file.GetID()] = file
			return file
		}
	}

	return nil
}

// GenerateIndex for search queries
func (zs *Handler) GenerateIndex(libPath string, skipIndexing bool) error {
	s := uint32(0)

	indexDB, err := NewIndexDB(libPath)
	if err != nil {
		return err
	}

	// Create index for all files
	for i := range zs.files {
		file := &zs.files[i]
		// Set index file
		_, fname := filepath.Split(file.Path)
		fname = "." + fname + ".ix"
		file.IndexFile = filepath.Join(libPath, fname)

		// Check file validation
		ok, err := indexDB.CheckFile(file.IndexFile)
		if err != nil {
			return err
		}
		// Skip file if index is still valid
		if ok {
			Log.Infof("Index for %s exists", file.Filename())
			continue
		}

		var size uint32
		if !skipIndexing {
			// Create new Index file
			f, err := os.OpenFile(file.IndexFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}

			// Generate index
			size, err = file.generateFileIndex(f)
			if err != nil {
				return err
			}

			f.Close()
		} else {
			Log.Warn("Skipping Index", file.Filename())
		}

		// Add index to DB
		err = indexDB.AddIndexFile(file.IndexFile, file.GetID())
		if err != nil {
			if err == ErrAlreadyInDB {
				continue
			}
			return err
		}

		s += size
	}

	if s > 0 {
		Log.Infof("Generated index size: %dMB\n", s/1000/1000)
	}

	return nil
}
