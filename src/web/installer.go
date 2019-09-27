package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"archive/zip"

	"github.com/labstack/echo"
)

const (
	installerVersion = "0.6.0"
)

type InstallerVersion bool

var JAR InstallerVersion = false
var EXE InstallerVersion = true

var installerJar []byte
var installerExe []byte

var ready = make(chan struct{})

func (version InstallerVersion) getEXT() string {
	if version == JAR {
		return "jar"
	} else {
		return "exe"
	}
}
func (version InstallerVersion) getURL() string {
	return "https://github.com/ImpactDevelopment/Installer/releases/download/" + installerVersion + "/installer-" + installerVersion + "." + version.getEXT()
}

func (version InstallerVersion) fetchFile() ([]byte, error) {
	url := version.getURL()
	fmt.Println("Downloading", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	fmt.Println("Finished downloading", url, "length is", len(data))
	return data, err
}

func (version InstallerVersion) incrementGithubDownloadCountButDontActuallyUseTheirS3Bandwidth() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(version.getURL())

	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode != 302 {
		fmt.Println("GitHub did not accept the request")
	}
}

func init() {
	// fetch the files on startup, but don't block init on it :brain:
	go startup()
}

func startup() {
	var err error
	installerJar, err = JAR.fetchFile()
	if err != nil {
		panic(err)
	}

	_, err = installerJarReader()
	if err != nil {
		panic(err)
	}

	installerExe, err = EXE.fetchFile()
	if err != nil {
		panic(err)
	}
	sanityCheck()

	fmt.Println("Initialized")
	go func() {
		for {
			ready <- struct{}{} // we are ready from now on
		}
	}()
}

func exeHeaderLen() int {
	return len(installerExe) - len(installerJar)
}

func exeHeader() []byte {
	return installerExe[:exeHeaderLen()]
}

func sanityCheck() {
	for i := 0; i < len(installerJar); i++ {
		if installerJar[i] != installerExe[exeHeaderLen()+i] {
			panic("invalid installer files")
		}
	}
}

func installerJarReader() (*zip.Reader, error) {
	return zip.NewReader(bytes.NewReader(installerJar), int64(len(installerJar)))
}

func awaitStartup() { // blocks and only returns once startup is done
	<-ready
}

func extractOrGenerateCID(c echo.Context) string {
	cid := extractTrackyTracky(c)
	if cid != "" {
		return cid
	}
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func extractTrackyTracky(c echo.Context) string {
	cookie, err := c.Cookie("_ga")
	if err != nil {
		return ""
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 4 {
		return ""
	}
	return parts[2] + "." + parts[3]
}

func installerForJar(c echo.Context) error {
	return installer(c, JAR)
}

func installerForExe(c echo.Context) error {
	return installer(c, EXE)
}

func analytics(cid string, version InstallerVersion) {
	data := url.Values{}
	data.Set("v", "1")
	data.Set("t", "event")
	data.Set("tid", "UA-143397381-1")
	data.Set("cid", cid)
	data.Set("ds", "backend")
	data.Set("ec", "installer")
	data.Set("ea", "download")
	data.Set("el", version.getEXT())

	req, _ := http.NewRequest("POST", "https://www.google-analytics.com/collect", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("Analytics error", err)
	}
	if resp.StatusCode != 200 {
		fmt.Println("Analytics bad status code", resp.StatusCode)
		data, err := ioutil.ReadAll(resp.Body)
		fmt.Println(err)
		fmt.Println(string(data))
	}
	fmt.Println("Analytics success")
}

func installer(c echo.Context, version InstallerVersion) error {
	awaitStartup() // in case we get an early request, block until startup is done
	reader, err := installerJarReader()
	if err != nil {
		panic(err)
	}

	res := c.Response()
	header := res.Header()
	header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	header.Set(echo.HeaderContentDisposition, "attachment; filename=ImpactInstaller-"+installerVersion+"."+version.getEXT())
	header.Set("Content-Transfer-Encoding", "binary")
	header.Set("Cache-Control", "max-age=0")
	res.WriteHeader(http.StatusOK)

	if version == EXE {
		_, err := res.Write(exeHeader())
		if err != nil {
			return err
		}
	}

	zipWriter := zip.NewWriter(res)
	defer zipWriter.Close()
	for _, file := range reader.File {
		entryWriter, err := zipWriter.Create(file.Name)
		if err != nil {
			return err
		}
		entryReader, err := file.Open()
		if err != nil {
			return err
		}
		defer entryReader.Close()
		_, err = io.Copy(entryWriter, entryReader)
		if err != nil {
			return err
		}
	}
	cid := extractOrGenerateCID(c)
	writer, err := zipWriter.Create("impact_cid.txt")
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(cid))
	if err != nil {
		return err
	}
	go analytics(cid, version)
	go version.incrementGithubDownloadCountButDontActuallyUseTheirS3Bandwidth()

	return nil
}
