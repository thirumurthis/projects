### Load archived image in local to target repo

Recently had to work on publishing the locally archived image to Nexus Sonatype artifactory. 


Pre-requsites:
 - Nexus repository accessible


To save the image locally we can use docker cli

```sh
docker pull nginx
docker pull busybox

docker save -o local-image.tar nginx busybox

gzip local-image.tar
```

We use Go lang containerregistry library crane to push the image to private registry. This doesn't require docker client.

- The code uses flag lib to pass CLI argument, the image archive tar.gz is passed to the code.
- The gzip is uncompressed to tar file and it is passed to the tarball library to load the manifest in the tar archive.
- The input archive image is passed to the code using -image-tar flag
- The -list flag will print the image info from the tar manifest.
- The -target-repo will take the target repo with the host. The local nexus was deployed in `nexus.local/local-docker`
- The crane library push the image to artifactory, it support multi image archive. In this case we have busybox and nginx image tar.gz file.

```go
package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
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

func CheckImageArchiveTarGzAndExtractTarToDestDir(imageTarGz, tarDestDir string) (*string, error) {
	gzSuffix := ".gz"
	destTarFilePathWithName := ""
	if strings.HasSuffix(imageTarGz, gzSuffix) {
		// get the tar file name from the input archive tar.gz
		tarFileName := GetTarPartFromTarGzFileName(imageTarGz)
		//get current working directory
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Error occurred getting cwd %v", err) // Use log.Fatal to handle errors gracefully
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

		slog.Debug(fmt.Sprintf("Image archive tar.gz file path name: %s", imageTarGz))
		destTarFilePathWithName = filepath.Join(tarDestDir, tarFileName)
		slog.Debug(fmt.Sprintf("Destination tar file path name: %s", destTarFilePathWithName))
		tarErr := ExtractTarGz(imageTarGz, destTarFilePathWithName, false)
		if tarErr != nil {
			return nil, fmt.Errorf("error occurred during tar file creation %v", tarErr)
		}
		slog.Info(fmt.Sprintf("Untarred file successfully. Path: %s", destTarFilePathWithName))
	} else {
		return nil, fmt.Errorf("image archive is not tar.gz format")
	}
	return &destTarFilePathWithName, nil
}

// to save the image in linux/ wsl
// $ cd input
// $ docker save -o local-image.tar nginx busybox
// $ gzip local-image.tar
// $ go run ./cmd/image-loader/ -u $USER -p $PASWD -image-tar ./input/local-image.tar.gz \
// -target-repo nexus.local/local-docker -list
func main() {

	username := flag.String("u", "", "artifactory username")
	password := flag.String("p", "", "artifactory password")
	insecure := flag.Bool("insecure", true, "enable when using TLS, insecure by default")
	imageTarGz := flag.String("image-tar", "", "the image archive in tar.gz saved locally, example using docker save -o image.tar nginx busybox; gzip image.tar")
	targetHostRepo := flag.String("target-repo", "", "the target host and repo of artifactory to publish image, example nexus.local/local-docker")
	verbosity := flag.Int("v", 3, "log level 1: error, 2. warn, 3. info, 4. debug")

	printManifest := flag.Bool("list", false, "set this flag to list images from manfiest within image archive")
	sc := flag.Bool("skip-publish", false, "set this flag to skip publishing image to target artifactory and only list the image from manifest, ")

	flag.Parse()

	setLogLevel(*verbosity)

	if *imageTarGz == "" {
		slog.Error("Image archive tar.gz is required, use -image-tar file-name.tar.gz")
		os.Exit(1)
	}

	if !strings.HasSuffix(*imageTarGz, ".gz") {
		slog.Error("Supports only tar.gz format")
		os.Exit(1)
	}

	if *targetHostRepo == "" {
		slog.Error("Target Host repo is required, use -target-repo <host>/<repo>")
		os.Exit(1)
	}

	_, fileError := FileExists(*imageTarGz)

	if fileError != nil {
		slog.Error(fmt.Sprintf("Image tar.gz file doesn't exists %v", fileError))
		os.Exit(1)
	}

	tarDestDir := "output"

	untarFilePath, extractError := CheckImageArchiveTarGzAndExtractTarToDestDir(*imageTarGz, tarDestDir)

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

func pathOpener(path string) tarball.Opener {
	return func() (io.ReadCloser, error) {
		return os.Open(path)
	}
}
```


The archived input image tar.gz is stored in input directory, to execute the go code


```sh
go run ./cmd/image-loader/ -u $ADMIN -p $PASSWRD -image-tar ./input/local-image.tar.gz \
-target-repo nexus.local/local-docker -list
```

- Output

```sh
time=2026-03-07T22:02:57.177-08:00 level=INFO msg="Untarred file successfully. Path: ..image-loader\\output\\local-image.tar"
Image info from manifest: nginx:alpine
Image info from manifest: nginx:latest
Image info from manifest: busybox:latest
time=2026-03-07T22:02:57.724-08:00 level=INFO msg="Successfully published image to repo" targetRepoImageUrl=nexus.local/local-docker/nginx:alpine
time=2026-03-07T22:02:57.866-08:00 level=INFO msg="Successfully published image to repo" targetRepoImageUrl=nexus.local/local-docker/nginx:latest
time=2026-03-07T22:02:57.953-08:00 level=INFO msg="Successfully published image to repo" targetRepoImageUrl=nexus.local/local-docker/busybox:latest
completed
```

Nexus artifactory before executing the code

<img width="1591" height="1010" alt="image" src="https://github.com/user-attachments/assets/54c5e621-03b1-4125-8bbf-2c6aeef4a553" />

After executing the command we could see the images are pushed

<img width="1357" height="1113" alt="image" src="https://github.com/user-attachments/assets/65b0e37a-aacb-4c53-856a-dfb4658c2bee" />
