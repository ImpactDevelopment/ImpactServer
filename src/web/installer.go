package web

import (
	"bytes"
	"errors"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"archive/zip"

	"github.com/labstack/echo"
)

var installerVersion string

type InstallerVersion int

const (
	JAR InstallerVersion = iota
	EXE
)

type Entry struct { // can't use zip.Entry since that seeks within the input and decompresses on the fly (slow)
	name string
	data []byte
}

var installerEntries []Entry
var exeHeader []byte

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
	util.LogInfo("Downloading " + url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			util.LogWarn("Error closing body. " + err.Error())
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	util.LogSuccess("Finished downloading " + url + " length is " + string(len(data)))
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
		util.LogWarn("Unable to get version, " + err.Error())
		return // prevent fallthrough, lines below wouldn't work if err isn't nil
	}
	if resp.StatusCode != 302 {
		util.LogInfo("GitHub did not accept the request")
	}
}

func init() {
	installerVersion = os.Getenv("INSTALLER_VERSION")
	if installerVersion == "" {
		util.LogWarn("Installer version not specified, download will not work!")
		return
	}
	// fetch the files on startup, but don't block init on it :brain:
	go startup()
}

func startup() {
	installerJar, err := JAR.fetchFile()
	if err != nil {
		panic(err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(installerJar), int64(len(installerJar)))
	if err != nil {
		panic(err)
	}

	installerEntries = make([]Entry, 0)
	for _, file := range zipReader.File {
		entryReader, err := file.Open()
		if err != nil {
			panic(err)
		}
		defer entryReader.Close()
		data, err := ioutil.ReadAll(entryReader)
		if err != nil {
			panic(err)
		}
		installerEntries = append(installerEntries, Entry{
			name: file.Name,
			data: data,
		})
	}

	installerExe, err := EXE.fetchFile()
	if err != nil {
		panic(err)
	}

	exeHeaderLen := len(installerExe) - len(installerJar)
	for i := 0; i < len(installerJar); i++ {
		if installerJar[i] != installerExe[exeHeaderLen+i] {
			panic("invalid installer files")
		}
	}
	exeHeader = installerExe[:exeHeaderLen]

	util.LogSuccess("Initialized")
	go func() {
		for {
			ready <- struct{}{} // we are ready from now on
		}
	}()
}

func awaitStartup() { // blocks and only returns once startup is done
	<-ready
}

func extractOrGenerateCID(c echo.Context) string {
	cid := extractTrackyTracky(c)
	if cid != "" {
		return cid
	}
	uuid, err := uuid.NewUUID()
	if err != nil {
		panic(err) // happens when system clock is not set or something dummy like that
	}
	return uuid.String()
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

func analytics(cid string, version InstallerVersion, c echo.Context) {
	data := url.Values{}
	data.Set("v", "1")
	data.Set("t", "event")
	data.Set("tid", "UA-143397381-1")
	data.Set("cid", cid)
	data.Set("ds", "backend")
	data.Set("ec", "installer")
	data.Set("ea", "download")
	data.Set("el", version.getEXT())
	data.Set("ua", c.Request().UserAgent())

	forward := strings.Split(c.Request().Header.Get("X-FORWARDED-FOR"), ",")[0]
	if forward != "" {
		data.Set("uip", forward)
	}

	req, _ := http.NewRequest("POST", "https://www.google-analytics.com/collect", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("User-Agent", c.Request().UserAgent())

	resp, err := (&http.Client{}).Do(req)
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			util.LogWarn("Unable to close body " + err.Error())
		}
	}()
	if err != nil {
		util.LogWarn("Analytics error" + err.Error())
		return // resp.StatusCode will be nil if err != nil
	}
	if resp.StatusCode != 200 {
		util.LogWarn("Analytics bad status code " + string(resp.StatusCode))
		data, err := ioutil.ReadAll(resp.Body)
		util.LogWarn(err)
		util.LogWarn(data)
	}
}

func makeEntry(zipWriter *zip.Writer, name string, exe bool) (io.Writer, error) {
	// make an entry with a valid last modified time so as to not crash java 12 reeee
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	if !exe {
		header.Modified = time.Now()
	}
	return zipWriter.CreateHeader(header)
}

func installer(c echo.Context, version InstallerVersion) error {
	if installerVersion == "" {
		return errors.New("Installer version not specified")
	}
	awaitStartup() // in case we get an early request, block until startup is done

	referer := c.Request().Referer()
	if referer != "" && !strings.HasPrefix(referer, "https://impactclient.net/") && !strings.Contains(referer, "brady-money-grubbing-completed") {
		util.LogInfo("BLOCKING referer " + referer)
		return echo.NewHTTPError(http.StatusUnauthorized, "no hotlinking >:(")
	}

	res := c.Response()
	header := res.Header()
	header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	header.Set(echo.HeaderContentDisposition, "attachment; filename=ImpactInstaller-"+installerVersion+"."+version.getEXT())
	header.Set("Content-Transfer-Encoding", "binary")
	res.WriteHeader(http.StatusOK)

	if version == EXE {
		_, err := res.Write(exeHeader)
		if err != nil {
			return err
		}
	}

	zipWriter := zip.NewWriter(res)
	defer zipWriter.Close()
	for _, entry := range installerEntries {
		entryWriter, err := makeEntry(zipWriter, entry.name, version == EXE)
		if err != nil {
			return err
		}
		_, err = entryWriter.Write(entry.data)
		if err != nil {
			return err
		}
	}
	cid := extractOrGenerateCID(c)
	writer, err := makeEntry(zipWriter, "impact_cid.txt", version == EXE)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(cid))
	if err != nil {
		return err
	}
	go analytics(cid, version, c)
	go version.incrementGithubDownloadCountButDontActuallyUseTheirS3Bandwidth()

	return nil
}
