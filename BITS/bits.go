/*
BAD2BEEF's Minimal BITS Server
https://github.com/bad2beef/Toolbox
This is a minimal BITS server. It makes no attempt at being complete, correct, or secure.
Reference: https://docs.microsoft.com/en-us/windows/win32/bits/bits-upload-protocol
*/

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var basePathLogs string = "logs"
var basePathBITS string = "bits"

func writeLog(address, session string, level int, message string) {
	timestamp := time.Now()
	year, month, day := timestamp.Date()
	logFile := fmt.Sprintf("%v/%d/%02d/%d-%02d-%d.log", basePathLogs, year, int(month), year, int(month), day)
	logPath := filepath.Dir(logFile)

	if _, err := os.Stat(logFile); err != nil {
		os.MkdirAll(logPath, os.ModePerm)
	}

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	logLine := fmt.Sprintf("%v %v %v %v\n", timestamp.Format(time.RFC3339), address, session, message)
	file.WriteString(logLine)
	fmt.Print(logLine)
}

func writeInternalServerError(httpWriter http.ResponseWriter, httpRequest *http.Request, session, message string) {
	writeLog(httpRequest.RemoteAddr, session, 4, message)
	httpWriter.WriteHeader(http.StatusInternalServerError)
	httpWriter.Write([]byte(fmt.Sprintf("500 %v", message)))
}

func randBetween(lower, upper int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(upper-lower+1) + lower
}

func bitsGetSessionID() string {
	return fmt.Sprintf("{%04X%04X-%04X-%04X-%04X-%04X%04X%04X}", randBetween(0, 65535), randBetween(0, 65535), randBetween(0, 65535), randBetween(16384, 20479), randBetween(32768, 49151), randBetween(0, 65535), randBetween(0, 65535), randBetween(0, 65535))
}

func bitsGetSession(httpWriter http.ResponseWriter, httpRequest *http.Request) (string, string, bool) {
	bitsSession, present := httpRequest.Header["Bits-Session-Id"]
	if !present {
		writeInternalServerError(httpWriter, httpRequest, bitsSession[0], "Missing BITS-Session-Id")
		return "", "", false
	}

	if matched, _ := regexp.MatchString(`^\{?[A-Z0-9]{8}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{12}\}?$`, bitsSession[0]); !matched {
		writeInternalServerError(httpWriter, httpRequest, bitsSession[0], "Invalid BITS-Session-Id")
		return "", "", false
	}

	bitsSessionPath := path.Join(basePathBITS, bitsSession[0][1:3], bitsSession[0][4:6], bitsSession[0][1:37])
	if _, err := os.Stat(bitsSessionPath); err != nil {
		writeInternalServerError(httpWriter, httpRequest, bitsSession[0], "BITS-Session-Id not found")
		return "", "", false
	}

	return bitsSession[0], bitsSessionPath, true
}

func httpHandlerBITS(httpWriter http.ResponseWriter, httpRequest *http.Request) {
	if !strings.EqualFold(httpRequest.Method, "BITS_POST") {
		writeInternalServerError(httpWriter, httpRequest, "-", "Invalid HTTP method")
		return
	}

	bitsAction, present := httpRequest.Header["Bits-Packet-Type"]
	if !present {
		writeInternalServerError(httpWriter, httpRequest, "-", "Missing BITS-Packet-Type")
		return
	}

	switch strings.ToLower(bitsAction[0]) {
	case "ping":
		httpWriter.Header().Add("BITS-Packet-Type", "Ack")
		httpWriter.Header().Set("Content-Length", "0")
		httpWriter.WriteHeader(http.StatusOK)
	case "create-session":
		bitsSession := bitsGetSessionID()
		bitsSessionPath := path.Join(basePathBITS, bitsSession[1:3], bitsSession[4:6], bitsSession[1:37])

		writeLog(httpRequest.RemoteAddr, bitsSession, 1, "Created session")

		if err := os.MkdirAll(bitsSessionPath, os.ModePerm); err != nil {
			writeInternalServerError(httpWriter, httpRequest, bitsSession, "Could not create session")
			return
		}

		httpWriter.Header().Add("BITS-Packet-Type", "Ack")
		httpWriter.Header().Add("BITS-Protocol", "{7df0354d-249b-430f-820d-3d2a9bef4931}")
		httpWriter.Header().Add("BITS-Session-Id", bitsSession)
		httpWriter.Header().Add("Accept-Encoding", "Identity")
		httpWriter.Header().Set("Content-Length", "0")
		httpWriter.WriteHeader(http.StatusOK)
	case "close-session":
		bitsSession, bitsSessionPath, valid := bitsGetSession(httpWriter, httpRequest)
		if !valid {
			return
		}

		fileContent, err := os.OpenFile(path.Join(bitsSessionPath, bitsSession[1:37]), os.O_RDONLY, 0600)
		if err != nil {
			writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Hash: %v", err))
		} else {
			defer fileContent.Close()

			hash := sha256.New()
			bufferRead := make([]byte, 16*1024)

			for {
				bytesRead, err := fileContent.Read(bufferRead)
				if bytesRead > 0 {
					_, _ = hash.Write(bufferRead[:bytesRead])
				}

				if err == io.EOF {
					break
				}

				if err != nil {
					writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Hash: %v", err))
					break
				}
			}

			fileHash, err := os.OpenFile(path.Join(bitsSessionPath, fmt.Sprintf("%v.Hash", bitsSession[1:37])), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Hash: %v", err))
			} else {
				defer fileHash.Close()

				fileChecksum := hash.Sum(nil)
				fileHash.WriteString(fmt.Sprintf("%x", fileChecksum))
				writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Hash: %x\n", fileChecksum))
			}
		}
		fallthrough
	case "cancel-session":
		bitsSession, _, valid := bitsGetSession(httpWriter, httpRequest)
		if !valid {
			return
		}

		httpWriter.Header().Add("BITS-Packet-Type", "Ack")
		httpWriter.Header().Add("BITS-Session-Id", bitsSession)
		httpWriter.Header().Set("Content-Length", "0")
		httpWriter.WriteHeader(http.StatusOK)
	case "fragment":
		bitsSession, bitsSessionPath, valid := bitsGetSession(httpWriter, httpRequest)
		if !valid {
			return
		}

		bitsFileName, present := httpRequest.Header["Content-Name"]
		if present {
			fileContentName, err := os.OpenFile(path.Join(bitsSessionPath, fmt.Sprintf("%v.Content-Name", bitsSession[1:37])), os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Content-Name: %v", err))
			} else {
				_, err := fileContentName.WriteString(bitsFileName[0])
				if err != nil {
					writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Content-Name: %v", err))
				}
			}
			fileContentName.Close()
		}

		bitsFileEncoding, present := httpRequest.Header["Content-Encoding"]
		if present {
			fileContentEncoding, err := os.OpenFile(path.Join(bitsSessionPath, fmt.Sprintf("%v.Content-Encoding", bitsSession[1:37])), os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Content-Encoding: %v", err))
			} else {
				_, err := fileContentEncoding.WriteString(bitsFileEncoding[0])
				if err != nil {
					writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Content-Encoding: %v", err))
				}
			}
			fileContentEncoding.Close()
		}

		bitsContentRange, present := httpRequest.Header["Content-Range"]
		if !present {
			writeInternalServerError(httpWriter, httpRequest, bitsSession, "Missing Content-Range")
			return
		}

		regexRange := regexp.MustCompile(`^bytes\s(\d+)\-(\d+)\/(\d+)$`)
		matchesRange := regexRange.FindStringSubmatch(bitsContentRange[0])

		fileContent, err := os.OpenFile(path.Join(bitsSessionPath, bitsSession[1:37]), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("File Create/Open: %v", err))
			writeInternalServerError(httpWriter, httpRequest, bitsSession, "Could not store content")
			return
		}
		defer fileContent.Close()

		fileContentRangeStart, _ := strconv.ParseInt(matchesRange[1], 10, 36)
		_, err = fileContent.Seek(fileContentRangeStart, 0)
		if err != nil {
			writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("File Seek: %v", err))
			writeInternalServerError(httpWriter, httpRequest, bitsSession, "Could not store content")
			return
		}

		body, err := ioutil.ReadAll(httpRequest.Body)
		written, err := fileContent.Write(body)
		if err != nil {
			writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("File Write: %v", err))
			writeInternalServerError(httpWriter, httpRequest, bitsSession, "Could not write content")
			return
		}

		writeLog(httpRequest.RemoteAddr, bitsSession, 1, fmt.Sprintf("Wrote %d bytes", written))

		fileContentRangeEnd, _ := strconv.ParseInt(matchesRange[2], 10, 36)
		httpWriter.Header().Add("BITS-Packet-Type", "Ack")
		httpWriter.Header().Add("BITS-Session-Id", bitsSession)
		httpWriter.Header().Add("BITS-Received-Content-Range", fmt.Sprintf("%d", fileContentRangeEnd+1))
		httpWriter.Header().Set("Content-Length", "0")
		httpWriter.WriteHeader(http.StatusOK)
	}
}

func doServer(listener, route string) {
	writeLog(listener, "-", 1, fmt.Sprintf("Starting lister with route %s", route))
	http.HandleFunc(route, httpHandlerBITS)
	if err := http.ListenAndServe(listener, nil); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	listener := flag.String("listen", ":8080", "HTTP listener address")
	route := flag.String("route", "/bits", "URL Route for BITS")
	pathBITS := flag.String("bits", "bits", "Base path for BITS transfers")
	pathLogs := flag.String("logs", "logs", "Base path for log files")
	flag.Parse()

	basePathBITS = *pathBITS
	basePathLogs = *pathLogs

	doServer(*listener, *route)
}
