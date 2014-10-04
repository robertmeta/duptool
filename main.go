package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type keepfile struct {
	path   string
	sha256 []byte
}

var keepfiles []keepfile
var removalList []string

func visit(path string, f os.FileInfo, err error) error {
	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal("visit1", err)
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("visit2", err)
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("visit3", err)
	}
	if !fi.IsDir() {
		hash := mustGetSHA256(file)
		for _, v := range keepfiles {
			if bytes.Equal(v.sha256, hash) {
				log.Println("Found Duplicate (removing): ", path)
				addAKA(v, path)
				return nil
			}
		}
		keepfiles = append(keepfiles, keepfile{path, hash})
	}
	return nil
}

func addAKA(k keepfile, dup string) {
	newPath := k.path + ".aka"
	akaFile, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("addAKA1", err)
	}
	defer akaFile.Close()
	_, err = akaFile.WriteString(dup + "\n")
	if err != nil {
		log.Fatal("addAKA2", err)
	}
	removalList = append(removalList, dup)
}

func mustGetSHA256(iw io.Reader) []byte {
	h := sha256.New()
	_, err := io.Copy(h, iw)
	if err != nil {
		log.Fatal("getSHA256", err)
	}
	v := h.Sum(nil)
	return v
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root, visit)
	if err != nil {
		log.Fatal("main", err)
	}
	for _, v := range removalList {
		err = os.Remove(v)
		if err != nil {
			log.Fatal("main2", err)
		}
	}
	fmt.Printf("Files deduped: %d", len(removalList))
}
