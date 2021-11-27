package rjio2

import (
	"fmt"
	"io/ioutil"

	"github.com/rs/zerolog/log"
)

type Storage interface {
	GetTemplate(filename string)
	GetOPML(filename string)
	StoreRSS(filename string)
}

type FileStorage struct {
	prefix string
}

func NewFileStorage(prefix string) *FileStorage {
	return &FileStorage{
		prefix: prefix,
	}
}

func (storage FileStorage) GetTemplate(filename string) string {
	target := storage.getTargetFilename(filename)
	bytes, err := ioutil.ReadFile(target)
	if err != nil {
		log.Fatal().Msgf("Cannot read source file, error=%v", err)
		panic(err)
	}

	return string(bytes)
}

func (storage FileStorage) StoreRSS(filename string, str string) {
	target := storage.getTargetFilename(filename)
	err := ioutil.WriteFile(target, []byte(str), 0644)
	if err != nil {
		log.Fatal().Msgf("Cannot write file, error=%v", err)
		panic(err)
	}
}

func (storage FileStorage) getTargetFilename(str string) string {
	return fmt.Sprintf("%s/%s", storage.prefix, str)
}
