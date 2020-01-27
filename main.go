package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func computeHash(img string, hashAlgorithm contrib.ImgHashBase) string {
	image := gocv.IMRead(img, gocv.IMReadColor)
	if image.Empty() {
		log.Printf("cannot read image %s\n", img)
		return ""
	}
	defer image.Close()
	hash := gocv.NewMat()
	defer hash.Close()
	hashAlgorithm.Compute(image, &hash)
	if hash.Empty() {
		log.Printf("error computing hash for %s\n", img)
		return ""
	}
	return fmt.Sprintf("%x", hash.ToBytes())

}

func generator(jobs chan string, files []string) {
	generators.Add(1)
	defer generators.Done()
	for _, file := range files {
		jobs <- file
	}
}

func worker(jobs chan string, done chan jobDone, folder string) {
	workers.Add(1)
	defer workers.Done()
	for {
		file, keepWorking := <-jobs
		if keepWorking {
			hashes := map[string]string{}
			for _, hashAlgorithm := range hashAlgorithms {
				hash := computeHash(folder+file, hashAlgorithm)
				name := strings.TrimPrefix(
					fmt.Sprintf("%T", hashAlgorithm),
					"contrib.",
				)
				hashes[name] = hash
			}
			newImage := NewImage(folder + file)
			hashes["Signature"] = newImage.Signature()
			done <- jobDone{file: file, hashes: hashes}
		} else {
			log.Printf("all jobs have been processed")
			return
		}
	}
}

func merger(done chan jobDone) {
	mergers.Add(1)
	defer mergers.Done()
	i := 1
	for job := range done {
		log.Printf(
			"(%.2f %%) [%d/%d] processing %s",
			float64(i*100)/float64(total),
			i,
			total,
			job.file,
		)
		images[job.file] = job.hashes
		i++
	}
}

type jobDone struct {
	file   string
	hashes map[string]string
}

var total = 0
var mergers sync.WaitGroup
var workers sync.WaitGroup
var generators sync.WaitGroup
var images = map[string]map[string]string{}
var hashAlgorithms = []contrib.ImgHashBase{
	//contrib.PHash{},
	//contrib.AverageHash{},
	//contrib.BlockMeanHash{},
	//contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1},
	//contrib.ColorMomentHash{},
	//contrib.NewMarrHildrethHash(),
	//contrib.NewRadialVarianceHash(),
}

func main() {
	extension := flag.String("extension", "", "file extension")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Printf("you must specify at least one folder.")
		return
	}
	nWorker := 4
	jobs := make(chan string, nWorker+1)
	done := make(chan jobDone, nWorker+1)

	for _, folder := range args {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}

		validFiles := []string{}
		for _, f := range files {
			if !f.IsDir() {
				if strings.HasSuffix(f.Name(), *extension) {
					validFiles = append(validFiles, f.Name())
				}
			}
		}
		total = len(validFiles)
		go generator(jobs, validFiles)
		go merger(done)
		for n := 1; n <= nWorker; n++ {
			log.Printf("starting worker %d", n)
			go worker(jobs, done, folder)
		}
		var sleep time.Duration = 5
		generators.Wait()
		log.Printf("all generators for folder '%s' are done... waiting %d seconds before continuing...", folder, sleep)
		time.Sleep(sleep * time.Second)
		close(jobs)
		workers.Wait()
		log.Printf("all workers for folder '%s' are done... waiting %d seconds before continuing...", folder, sleep)
		time.Sleep(sleep * time.Second)
		close(done)
		mergers.Wait()
		log.Printf("all mergers for folder '%s' are done... waiting %d seconds before continuing...", folder, sleep)
		time.Sleep(sleep * time.Second)
	}

	jsonResult, err := json.MarshalIndent(images, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", jsonResult)
}
