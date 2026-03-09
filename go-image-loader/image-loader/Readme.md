### Load archived image in local to target repo

Recently was exploring options to save container images locally as tar file and load programmatically using Golang to locally deployed Nexus Sonatype artifactory.

In this case instead of using the docker client or docker desktop, have used Go lang containerregistry library which supports to handle OCI image artifacts. The container image is pushed using docker which includes the multi image tar.gz archive in specific path, the code takes the image tag to downloads the container from artifactory and uses the Go library to extract the archive tar.gz to local machine. The extracted tar.gz can be used to load back to the artifactory, where each image in the archive tar.gz is loaded separately

Pre-requisites:
 - Nexus repository - in the demonstration have deployed the Nexus Sonatype in docker with Nginx reverse proxy with self signed certificate.
 - Go lang SDK installed

To save the image locally as tar we can use below command

```sh
docker pull nginx
docker pull busybox

docker save -o local-image.tar nginx:latest nginx:alpine busybox
```

Below command will compress the tar to gz format  

```sh
gzip local-image.tar
```

The Go lang containerregistry package used to load the image. The crane library is used since the tar archive used multiple images. If the tar includes just single image, we could simply use the containerregistry package to push the image to private registry. This doesn't require docker client. 

The Go lang code implements two scenarios
  1. With the local image archive in tar.gz format available, we can use below flags to load the image to private artifactory

```sh
 USER=xxx; PASSWRD=yyyy; go run ./cmd/image-loader/ -u $USER -p $PASSWRD -load-image-archive -image-archive-tar-to-load ./output/extract/local-image.tar.gz  -target-repo nexus.local/local-docker -list 
```

  2. With container created with the tar.gz image archive loaded to `/app/input/local-image.tar.gz`, we can use below flags to extract the image.

```sh
 USER=xxx; PASSWRD=yyyy; go run ./cmd/image-loader/ -extract-image-archive -image-tag-to-extract-archive nexus.local/local-docker/my-image:1.0.0  -cert-path nexus.local.crt -u $USER -p $PASSWRD -archive-path-in-container app/input/local-image.tar.gz
```

- If Go application executable is created using `go build ./cmd/image-loader/`, the image-loader.exe can be used with flags to load or extract image.

- For the load container image from the image tar.gz archive, the tar.gz is uncompressed to tar file. The tar file is passed to the tarball library which extracts the manfiest and image list is printed. The repoTags from the manifest is used to create the target repo using the input passed to the application. With the docker.io image, the archived manifest image tag looks like `nginx:latest`, etc. The target repo value is prefixed.
- For the extracting image within the container image from the Nexus artifactory, the image is loaded using remote library. The image tar is used to extract the content of the image. For each path in the image the file path is compared with the input archive path in the container and content is downloaded. 

<code>

The image tar.gz archived input is stored in input directory, execute the Go code with below command

```sh
go run ./cmd/image-loader/ -u $ADMIN -p $PASSWRD -image-tar ./input/local-image.tar.gz \
-target-repo nexus.local/local-docker -list
```

Output of the command

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

Go application command to extract the image archive tar.gz stored within container from artifactory.

```sh
 USR=xxx; PASSWRD=yyy; ./image-loader.exe -u $USR -p $PASSWRD -extract-image-archive -image-tag-to-extract-archive nexus.local/local-docker/my-image:1.0.0 -cert-path "/c/thiru/edu/go/apps/image-handler/nexus.local.crt" -archive-path-in-container app/input/local-image.tar.gz 
```

Output of the above command looks like below

```sh 
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="username: *****"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="password: ******"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="load-image-archive: false"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="image-archive-tar-to-load: "
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="target-repo: "
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="list: false"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="skip-publish: false"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="extract-image-archive: true"
time=2026-03-08T17:14:55.777-07:00 level=INFO msg="image-tag-to-extract-archive: nexus.local/local-docker/my-image:1.0.0"
time=2026-03-08T17:14:55.778-07:00 level=INFO msg="archive-path-in-container: app/input/local-image.tar.gz"
time=2026-03-08T17:14:55.778-07:00 level=INFO msg="cert-path: C:/thiru/edu/go/apps/image-handler/nexus.local.crt"
time=2026-03-08T17:14:55.778-07:00 level=INFO msg="insecure: true"
time=2026-03-08T17:14:55.778-07:00 level=INFO msg="Flow to extract image archive (tar.gz) from container"
time=2026-03-08T17:14:56.196-07:00 level=INFO msg="Completed extracting the image archive from container..."
```


Nexus UI before executing the Go code

<img width="1591" height="1010" alt="image" src="https://github.com/user-attachments/assets/54c5e621-03b1-4125-8bbf-2c6aeef4a553" />


Below is the Nexus UI after executing the Go code to load the image archive which load three images, nginx:latest, nginx:alpine, busybox:latestx

<img width="1357" height="1113" alt="image" src="https://github.com/user-attachments/assets/65b0e37a-aacb-4c53-856a-dfb4658c2bee" />


To configure the Nexus to be accessed as anonymous user, create the role, the screen looks like below 

<img width="1903" height="1360" alt="image" src="https://github.com/user-attachments/assets/b8d4750f-17f3-4ec0-a090-dabc9cd75758" />

Provide a Role id name and Role name

<img width="926" height="1452" alt="image" src="https://github.com/user-attachments/assets/970a2754-cf81-4e28-9249-b0017ec0427f" />

