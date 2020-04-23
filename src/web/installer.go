package web

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/util"

	"github.com/google/uuid"

	"archive/zip"

	"github.com/labstack/echo/v4"
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

var ready bool

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

	request, err := util.GetRequest(url)
	if err != nil {
		return nil, err
	}
	response, err := request.Do()
	if err != nil {
		return nil, err
	}
	if !response.Ok() {
		return nil, errors.New("Installer download status not OK")
	}

	data := response.Body
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
	installerVersion = os.Getenv("INSTALLER_VERSION")
	if installerVersion == "" {
		fmt.Println("WARNING: Installer version not specified, download will not work!")
		return
	}
	// fetch the files on startup, but don't block init on it :brain:
	go startup()
}

func startup() {
	// Download the two files in parallel
	jarCh := make(chan []byte)
	exeCh := make(chan []byte)
	go func() { jarCh <- downloadUntilSuccess(JAR, 5*time.Minute) }()
	go func() { exeCh <- downloadUntilSuccess(EXE, 5*time.Minute) }()
	var installerJar = <-jarCh
	var installerExe = <-exeCh

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

	exeHeaderLen := len(installerExe) - len(installerJar)
	for i := 0; i < len(installerJar); i++ {
		if installerJar[i] != installerExe[exeHeaderLen+i] {
			panic("invalid installer files")
		}
	}
	exeHeader = installerExe[:exeHeaderLen]

	fmt.Println("Initialized")
	ready = true // we are ready from now on
}

// download the installer, if an error occurs keep trying again after duration (blocking)
func downloadUntilSuccess(version InstallerVersion, duration time.Duration) []byte {
	var attempts int
	exe, err := version.fetchFile()
	for err != nil {
		attempts++
		fmt.Fprintf(os.Stderr, "Error downloading %s Installer after %d attempts: %s\n", version.getEXT(), attempts, err.Error())
		time.Sleep(duration)
		exe, err = version.fetchFile()
	}
	return exe
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
	form := map[string]string{
		"v":   "1",
		"t":   "event",
		"tid": "UA-143397381-1",
		"cid": cid,
		"ds":  "backend",
		"ec":  "installer",
		"ea":  "download",
		"el":  version.getEXT(),
		"ua":  c.Request().UserAgent(),
	}
	forward := util.RealIPIfUnambiguous(c)
	if forward != "" {
		form["uip"] = forward
	}
	req, err := util.FormRequest("https://www.google-analytics.com/collect", form)
	if err != nil {
		fmt.Println("Analytics failed to build request", err)
		return
	}
	req.SetHeader("User-Agent", c.Request().UserAgent())

	resp, err := req.Do()
	if err != nil {
		fmt.Println("Analytics error", err)
		return
	}
	if !resp.Ok() {
		fmt.Println("Analytics bad status code", resp.Status())
		data := resp.String()
		fmt.Println(err)
		fmt.Println(data)
	}
}

func makeEntry(zipWriter *zip.Writer, entryName string, entry []byte, version InstallerVersion) error {
	// make an entry with a valid last modified time so as to not crash java 12 reeee
	header := &zip.FileHeader{
		Name:   entryName,
		Method: zip.Deflate,
	}
	switch version {
	case EXE: // Don't set modified time for EXE versions
	default:
		header.Modified = time.Now()
	}

	// make the entry
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(entry))
	if err != nil {
		return err
	}
	return nil
}

func installer(c echo.Context, version InstallerVersion) error {
	if installerVersion == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Installer version not specified")
	}

	referer := c.Request().Referer()
	if referer != "" && !strings.HasPrefix(referer, util.GetServerURL().String()) && !strings.Contains(referer, "brady-money-grubbing-completed") {
		fmt.Println("BLOCKING referer", referer)
		return echo.NewHTTPError(http.StatusUnauthorized, "no hotlinking >:(")
	}

	if !ready {
		c.Response().Header().Set("Retry-After", "120")
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Installer download not ready yet, please try again later")
	}

	nightlies := c.QueryParam("nightlies") == "1" || c.QueryParam("nightlies") == "true"
	filename := "Impact"
	if nightlies {
		filename += "Nightly"
	}
	filename += "Installer-" + installerVersion + "." + version.getEXT()

	res := c.Response()
	header := res.Header()
	header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	header.Set(echo.HeaderContentDisposition, "attachment; filename="+filename)
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
		err := makeEntry(zipWriter, entry.name, entry.data, version)
		if err != nil {
			return err
		}
	}
	if nightlies {
		const properties = "# Enable nightly builds\n" +
			"noGPG = true\n" +
			"prereleases = true\n"
		err := makeEntry(zipWriter, "default_args.properties", []byte(properties), version)
		if err != nil {
			return err
		}
	}
	cid := extractOrGenerateCID(c)
	err := makeEntry(zipWriter, "impact_cid.txt", []byte(cid), version)
	if err != nil {
		return err
	}
	go analytics(cid, version, c)
	go version.incrementGithubDownloadCountButDontActuallyUseTheirS3Bandwidth()

	return nil
}
