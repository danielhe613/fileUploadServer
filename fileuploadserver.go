// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	atomic "sync/atomic"
	//	"github.com/prometheus/common/log"
)

//TmpSeqNo is the sequence number for temporary files naming.
var TmpSeqNo int64 = 1

//FilePrefix is the uploaded file name's path.
var FilePrefix = "./uploaded"

//IsNotExist checks if the directory named with given path does exist or not. If not, return true.
func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func main() {

	//Check if export folder exists or not
	if IsNotExist("uploaded") {
		log.Fatalln("./uploaded folder for exported data files does NOT exist! Please create and make it writable.")
	}

	//Launch HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		go func() {
			//Check HTTP Method, and Date and Signature in the future
			if r.Method != "PUT" {
				log.Println("The HTTP method is NOT PUT!")
				return
			}

			//		fmt.Println(r.RequestURI)

			filename := FilePrefix + r.RequestURI

			//Checks if filename exists or not
			f, err := os.Open(filename)
			if err == nil || !os.IsNotExist(err) {
				defer f.Close()
				log.Printf("File %s already uploaded!\n", filename)
				return
			}

			//Creates the temp file
			tmpFilename := FilePrefix + r.RequestURI + "." + string(atomic.AddInt64(&TmpSeqNo, 1))
			tmpFile, err := os.Create(tmpFilename)
			defer func() {
				if tmpFile != nil {
					tmpFile.Close()
				}
			}()
			if err != nil {
				log.Printf("Error %s occurred when creating tmp file!", err)
				w.Write([]byte("Error to create tmp file!\n"))
				return
			}

			//Stream copying
			io.Copy(tmpFile, r.Body)

			//Rename filename.uploading to filename
			err = os.Rename(tmpFilename, filename)
			if err != nil {
				log.Printf("Error %s occurred when renaming the tmp file!", err)
				w.Write([]byte("Error to rename the tmp file!\n"))
				return
			}

		}()
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
