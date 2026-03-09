package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func FileExists(filePath string) (bool, error) {

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return false, fmt.Errorf("File not exist or Error occured during file check")
}

func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func GetTarPartFromTarGzFileName(gzFilePath string) string {
	gzExtension := ".gz"

	if gzFilePath == "" || !strings.HasSuffix(gzFilePath, gzExtension) {
		fmt.Printf("file %s is not a gz file", gzFilePath)
		return ""
	}
	gzTarFilename := filepath.Base(gzFilePath)
	tarFileName := strings.TrimSuffix(gzTarFilename, gzExtension)
	return tarFileName
}

// if the extractTar content set the flag to true
func ExtractTarGz(src, dest string, extractTarContent bool) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	if extractTarContent {
		tr := tar.NewReader(gzr)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			target := filepath.Join(dest, header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				os.MkdirAll(target, 0755)
			case tar.TypeReg:
				os.MkdirAll(filepath.Dir(target), 0755)
				outFile, _ := os.Create(target)
				io.Copy(outFile, tr) // Uses io.Copy for data extraction
				outFile.Close()
			}
		}
	} else {
		tarOutFile, err := os.Create(dest)
		if err != nil {
			log.Fatalf("cannot create the tar file from the gz format")
		}
		defer tarOutFile.Close()
		_, err = io.Copy(tarOutFile, gzr)

		if err != nil {
			log.Fatal("error occured during copy to tar")
		}
		if err := tarOutFile.Close(); err != nil {
			log.Fatal("error closing the file")
		}
	}
	return nil
}

func pathOpener(path string) tarball.Opener {
	return func() (io.ReadCloser, error) {
		return os.Open(path)
	}
}

func setLogLevel(verbosity int) {

	var logLevel = new(slog.LevelVar)

	switch verbosity {
	case 1:
		logLevel.Set(slog.LevelError)
	case 2:
		logLevel.Set(slog.LevelWarn)
	case 3:
		logLevel.Set(slog.LevelInfo)
	case 4:
		logLevel.Set(slog.LevelDebug)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
}

func GenerateTargetRepo(imageTagFromTarManfiest, targetHostBaseRepo string) string {
	targetRepo := ""
	targetHostBaseRepo = strings.TrimSuffix(targetHostBaseRepo, "/")
	inputSlice := []string{targetHostBaseRepo, imageTagFromTarManfiest}
	slog.Debug("Image tag from the manfiest", "imageTagFromTarManifest", imageTagFromTarManfiest)
	slog.Debug("Target Host and repo", "targetHostRepo", targetHostBaseRepo)
	if strings.Contains(imageTagFromTarManfiest, "/") {
		splitImageUrlTagToParts := strings.Split(imageTagFromTarManfiest, "/")
		var slicetoAddNewTargetRepo []string
		slicetoAddNewTargetRepo = append(slicetoAddNewTargetRepo, targetHostBaseRepo)
		for _, part := range splitImageUrlTagToParts[1:] {
			slicetoAddNewTargetRepo = append(slicetoAddNewTargetRepo, part)
		}
		targetRepo = strings.Join(slicetoAddNewTargetRepo, "/")

	} else {
		targetRepo = strings.Join(inputSlice, "/")
	}
	slog.Debug("Final target image url", "targetImageUrlWithHost", targetRepo)
	return targetRepo
}

func CheckImageArchiveTarGzAndExtractTarToDestDir(imageArchiveTarGzfileToLoad, tarDestDir string) (*string, error) {
	gzSuffix := ".gz"
	destTarFilePathWithName := ""
	if strings.HasSuffix(imageArchiveTarGzfileToLoad, gzSuffix) {
		// get the tar file name from the input archive tar.gz
		tarFileName := GetTarPartFromTarGzFileName(imageArchiveTarGzfileToLoad)
		//get current working directory
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Error occurred getting cwd %v", err)
		}
		//create a new temp directory or output to extract the tar file
		tarDestDir := filepath.Join(dir, tarDestDir)

		// if the dest directory to untar doesn't exists create one
		if dirStatus, dirExistsErr := DirExists(tarDestDir); !dirStatus && dirExistsErr == nil {
			dirErr := os.MkdirAll(tarDestDir, os.ModePerm)

			if dirErr != nil {
				return nil, fmt.Errorf("error creating destination directory")
			}
		} else if dirExistsErr != nil {
			return nil, fmt.Errorf("tar destination dir check failed %s", tarDestDir)
		}

		slog.Debug(fmt.Sprintf("Image archive tar.gz file path name: %s", imageArchiveTarGzfileToLoad))
		destTarFilePathWithName = filepath.Join(tarDestDir, tarFileName)
		slog.Debug(fmt.Sprintf("Destination tar file path name: %s", destTarFilePathWithName))
		tarErr := ExtractTarGz(imageArchiveTarGzfileToLoad, destTarFilePathWithName, false)
		if tarErr != nil {
			return nil, fmt.Errorf("error occurred during tar file creation %v", tarErr)
		}
		slog.Info(fmt.Sprintf("Untarred file successfully. Path: %s", destTarFilePathWithName))
	} else {
		return nil, fmt.Errorf("image archive is not tar.gz format")
	}
	return &destTarFilePathWithName, nil
}

func CheckAndCreateDir(targetDestDir string) (*string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error occurred getting cwd %v", err)
	}
	//create a new temp directory or output to extract the tar file
	tarDestDir := filepath.Join(dir, targetDestDir)

	// if the dest directory to untar doesn't exists create one
	if dirStatus, dirExistsErr := DirExists(tarDestDir); !dirStatus && dirExistsErr == nil {
		dirErr := os.MkdirAll(tarDestDir, os.ModePerm)

		if dirErr != nil {
			return nil, fmt.Errorf("error creating destination directory")
		}
	} else if dirExistsErr != nil {
		return nil, fmt.Errorf("tar destination dir check failed %s", tarDestDir)
	}
	return &tarDestDir, nil
}

func extractFileFromImage(repoStr, username, password, archiveFilePathInImage,
	destFolderPathToCopyExtractArchive, certFilePath string, insecure bool) error {

	slog.Debug(fmt.Sprintf("Input - image tag to extract the archive from : %s", repoStr))
	slog.Debug(fmt.Sprintf("Input - image archive path image inside the container to extract from : %s", archiveFilePathInImage))
	slog.Debug(fmt.Sprintf("Input - image archiver to be extracted to host in this path : %s", destFolderPathToCopyExtractArchive))
	slog.Debug(fmt.Sprintf("Input - cert file path: %s", certFilePath))
	slog.Debug(fmt.Sprintf("Input - insecure option enabled ? %v", insecure))
	//Check if cert file exists
	_, certFileExistErr := FileExists(certFilePath)

	if certFileExistErr != nil {
		slog.Error(fmt.Sprintf("[%s] crt file doesn't exists", certFilePath))
		return certFileExistErr
	}

	//Load CA cert
	caCert, _ := os.ReadFile(certFilePath)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	//Create transport
	tlsConfig := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	//parse image reference
	ref, _ := name.ParseReference(repoStr)

	//Get the image from artifactory
	img, err := remote.Image(ref,
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password}),
		remote.WithTransport(tlsConfig))

	if username == "" && password == "" {
		slog.Info("Credentials not passed for extracting archive file from container")
		img, err = remote.Image(ref,
			remote.WithTransport(tlsConfig))
	}
	//
	if username == "" && password == "" && insecure {
		slog.Info("Credentials not passed, insecure mode enabled to extract archive file from container")
		img, err = remote.Image(ref)
	}

	if err != nil {
		return err
	}

	layers, err := img.Layers()
	if err != nil {
		return err
	}

	var fileData []byte
	dirName, _ := CheckAndCreateDir(destFolderPathToCopyExtractArchive)

	archiveFileTarGzFileNameParts := strings.Split(archiveFilePathInImage, "/")
	fileNameTarGzToCopyToHost := archiveFileTarGzFileNameParts[len(archiveFileTarGzFileNameParts)-1]
	slog.Debug(fmt.Sprintf("tar.gz file name from the input image archive path in container: %s", fileNameTarGzToCopyToHost))
	destFilePath := filepath.Join(*dirName, fileNameTarGzToCopyToHost)

	cleanedArchivePathInContainer := filepath.Clean("/" + archiveFilePathInImage)

	for _, layer := range layers {

		rc, err := layer.Uncompressed()

		if err != nil {
			return err
		}

		tr := tar.NewReader(rc)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			slog.Debug(fmt.Sprintf("Input image archive path in container: [%s], header name in tar: [%s]", cleanedArchivePathInContainer, "/"+hdr.Name))
			if filepath.Clean("/"+hdr.Name) == cleanedArchivePathInContainer {
				data, err := io.ReadAll(tr)
				rc.Close()
				if err != nil {
					return err
				}
				fileData = data
			}
		}
		rc.Close()
	}
	if fileData == nil {
		return fmt.Errorf("file %s not found in image", archiveFilePathInImage)
	}

	//fmt.Printf("filedata: %v", fileData)
	return os.WriteFile(destFilePath, fileData, 0644)
}

// to save the image in linux/ wsl
// $ cd input
// $ docker save -o local-image.tar nginx busybox
// $ gzip local-image.tar
// $ USER=xxxx; PASSWRD=xxxx;  go run ./cmd/image-loader/ -u $USER -p $PASSWRD -load-image-archive \
// -image-archive-tar-to-load ./output/extract/local-image.tar.gz  -target-repo nexus.local/local-docker -list -v 4
//
// To extract the tar.gz file from container without docker
// $ USER=xxxx; PASSWRD=xxxx; go run ./cmd/image-loader/ -extract-image-archive -image-tag-to-extract-archive nexus.local/local-docker/my-image:1.0.0 \
// -cert-path nexus.local.crt -u $USER -p $PASSWRD -archive-path-in-container app/input/local-image.tar.gz
func main() {

	username := flag.String("u", "", "artifactory username")
	password := flag.String("p", "", "artifactory password")
	insecure := flag.Bool("insecure", true, "enable when using TLS, insecure by default")
	imageArchiveTarGzfileToLoad := flag.String("image-archive-tar-to-load", "", "the image archive in tar.gz saved locally, example using docker save -o image.tar nginx busybox; gzip image.tar. This flag is passed along with -load-image-archive option")
	targetHostRepo := flag.String("target-repo", "", "the target host and repo of artifactory to publish image, example nexus.local/local-docker. This flag is passed along with -load-image-archive option")
	verbosity := flag.Int("v", 3, "Set on of the values for log level 1: error, 2. warn, 3. info (default), 4. debug")
	printManifest := flag.Bool("list", false, "settting this flag to list images from manfiest within image archive, using skip-publish flag will only list the image tag in image archive tar.gz. This flag is passed along with -load-image-archive option")
	sc := flag.Bool("skip-publish", false, "setting this flag will skip publishing image to target artifactory and only list the image from manifest.This flag is passed along with -load-image-archive option")
	loadImageArchiverToArtifactory := flag.Bool("load-image-archive", false, "Set this flag to load image archive (tar.gz) from host path. Only load archive or extract archive is supported the -extract-image-archive flag should not be provided when this option is passed")
	extractImageArchiveFromContainer := flag.Bool("extract-image-archive", false, "Set this flag to extract the image asrchive from container directly.Only load archive or extract archive is supported the -load-image-archive flag should not be provided when this option is passed")
	imageTagToExtractImageArchive := flag.String("image-tag-to-extract-archive", "", "image url to get the archive tar. This flag is passed along with -extract-image-archive option")
	imageArchivePathInContainer := flag.String("archive-path-in-container", "", "image archvie path in the container from where ther archive to be downloaded to host machine.  This flag is passed along with -extract-image-archive option")
	publicCertFile := flag.String("cert-path", "", "Artifactory public crt file, for slef-signed certificate export the .crt file from browser. This flag is passed along with -extract-image-archive option only")
	help := flag.Bool("help", false, "list the command usage option flags")
	flag.Parse()

	setLogLevel(*verbosity)

	if *username != "" {
		slog.Info("username: *****")
	}
	if *password != "" {
		slog.Info("password: ******")
	}
	slog.Info(fmt.Sprintf("load-image-archive: %t", *loadImageArchiverToArtifactory))
	slog.Info(fmt.Sprintf("image-archive-tar-to-load: %s", *imageArchiveTarGzfileToLoad))
	slog.Info(fmt.Sprintf("target-repo: %s", *targetHostRepo))
	slog.Info(fmt.Sprintf("list: %t", *printManifest))
	slog.Info(fmt.Sprintf("skip-publish: %t", *sc))

	slog.Info(fmt.Sprintf("extract-image-archive: %t", *extractImageArchiveFromContainer))
	slog.Info(fmt.Sprintf("image-tag-to-extract-archive: %s", *imageTagToExtractImageArchive))
	slog.Info(fmt.Sprintf("archive-path-in-container: %s", *imageArchivePathInContainer))
	slog.Info(fmt.Sprintf("cert-path: %s", *publicCertFile))

	slog.Info(fmt.Sprintf("insecure: %t", *insecure))

	//path in host machine where the tar should be extracted
	destFolderPathToCopyExtractArchive := "output/extract"
	if *help {
		flag.Usage()
		return
	}

	if *extractImageArchiveFromContainer && !*loadImageArchiverToArtifactory {
		slog.Info("Flow to extract image archive (tar.gz) from container")
		if *imageTagToExtractImageArchive == "" || *imageArchivePathInContainer == "" {
			slog.Error("The image url to extract the image url is required")
			slog.Error(fmt.Sprintf("image tag to extract from archive ? %s", *imageArchiveTarGzfileToLoad))
			slog.Error("The image archive path in container is required to extract")
			slog.Error(fmt.Sprintf("path of the image archive inside container ? %s", *imageArchivePathInContainer))
			os.Exit(1)
		} else {

			extactErr := extractFileFromImage(*imageTagToExtractImageArchive, *username,
				*password, *imageArchivePathInContainer, destFolderPathToCopyExtractArchive, *publicCertFile, false)
			if extactErr != nil {
				slog.Error(fmt.Sprintf("error occurred extracting tar - %v", extactErr))
			}
			slog.Info("Completed extracting the image archive from container...")
			return
		}
	}
	if *loadImageArchiverToArtifactory && !*extractImageArchiveFromContainer {
		slog.Info("Flow to load image archive (tar.gz) to target artifactory")
		if *imageArchiveTarGzfileToLoad == "" {
			slog.Error("Image archive tar.gz is required, use -image-tar file-name.tar.gz")
			os.Exit(1)
		}

		if !strings.HasSuffix(*imageArchiveTarGzfileToLoad, ".gz") {
			slog.Error("Supports only tar.gz format")
			os.Exit(1)
		}

		if *targetHostRepo == "" {
			slog.Error("Target Host repo is required, use -target-repo <host>/<repo>")
			os.Exit(1)
		}

		_, fileError := FileExists(*imageArchiveTarGzfileToLoad)

		if fileError != nil {
			slog.Error(fmt.Sprintf("Image tar.gz file doesn't exists %v", fileError))
			os.Exit(1)
		}

		//providedTarFile := *imageArchiveTarGzfileToLoad
		tarDestDir := "output"

		untarFilePath, extractError := CheckImageArchiveTarGzAndExtractTarToDestDir(*imageArchiveTarGzfileToLoad, tarDestDir)

		if extractError != nil {
			slog.Error("Exception occurred during untar process", "errorMessage", extractError)
		}
		slog.Debug(fmt.Sprintf("Untarred successfully to path: %s", *untarFilePath))

		manifests, err := tarball.LoadManifest(pathOpener(*untarFilePath))
		if err != nil {
			panic(err)
		}
		if *printManifest {
			for _, d := range manifests {
				for _, imageTag := range d.RepoTags {
					fmt.Printf("Image info from manifest: %s\n", imageTag)

				}
			}
			if *sc {
				return
			}

		}

		auth := &authn.Basic{
			Username: *username,
			Password: *password,
		}
		isUserPassProvided := false
		if *username != "" && *password != "" {
			isUserPassProvided = true

		}

		for _, descriptor := range manifests {
			for _, repoTag := range descriptor.RepoTags {
				tag, err := name.NewTag(repoTag)
				if err != nil {
					panic(err)
				}
				image, err := tarball.Image(pathOpener(*untarFilePath), &tag)
				if err != nil {
					panic(err)
				}
				slog.Debug("Repo tag from the manifest", "repoTag", repoTag)

				targetRepoToPublishImage := GenerateTargetRepo(repoTag, *targetHostRepo)

				if *insecure {
					if isUserPassProvided {
						slog.Debug("Publishing image to repo in insecure mode with credentials")
						err = crane.Push(image, targetRepoToPublishImage, crane.Insecure, crane.WithAuth(auth))
						slog.Info("Successfully published image to repo", "targetRepoImageUrl", targetRepoToPublishImage)
					} else {
						slog.Debug("Publishing image to repo in insecure mode with NO credentials")
						err = crane.Push(image, targetRepoToPublishImage, crane.Insecure)
						slog.Info("Successfully published image to repo", "targetRepoImageUrl", targetRepoToPublishImage)
					}

				} else {
					//only used when the public certificate is configured in golang code
					if isUserPassProvided {
						slog.Debug("Publishing image to repo with credentials")
						err = crane.Push(image, targetRepoToPublishImage, crane.WithAuth(auth))
						slog.Info("Successfully published image to repo", "targetRepoImageUrl", targetRepoToPublishImage)
					} else {
						slog.Debug("Publishing image to repo with NO credentials")
						err = crane.Push(image, targetRepoToPublishImage)
						slog.Info("Successfully published image to repo", "targetRepoImageUrl", targetRepoToPublishImage)
					}
				}
				if err != nil {
					slog.Error("Error occurred when publishing image")
					panic(err)
				}
			}

		}
		os.Remove(*untarFilePath)
		fmt.Println("completed")
	}
}
