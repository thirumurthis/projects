### Simple CLI built using Jbang to access the S3 compatible service with some basic operations

Pre-requistes
  - S3 compatible store is deployed, for development in my case used KinD deployed with Seaweedfs and to expose the S3 gateway as SSL, cert-manager and Apisix route were used.
  - Jbang CLI installed

The folder structure of the Jbang S3 Cli app

```
.
└── s3cliapp
    ├── app.java                <--- Entry point for app
    ├── application.yaml
    └── support
        ├── cliOptions.java
        ├── s3Handler.java
        └── s3Object.java
```

The Spring Boot, Pico Cli and S3 SDK dependencies are used to build this app. PicoCli dependency is used to handle option that can be passed to the java application. 

The picocli options can also read from the environment variables as well.

For example, execution below command will throw error message like below

```
> jbang s3cliapp\app.java
```

Output

```
$ jbang app.java 
[jbang] Building jar for app.java...
Missing required options: '--endpoint=<endpointUrl>', '--access-key=<accessKey>', '--secret-key=<secretKey>', '--operation=<operation>'
Usage: s3cli [-h] --access-key=<accessKey> [--bucket=<bucketName>]
             [--cert=<certPath>] [--content-type=<contentType>]
             --endpoint=<endpointUrl> [--file=<file>] --operation=<operation>
             [--region=<s3region>] --secret-key=<secretKey>
s3cli operations create and list buckets, upload file.
      --access-key=<accessKey>
                            use this option to pass access key, alternatively
                              S3_ACCESS_KEY env variable can also be used
      --bucket=<bucketName> use this option to pass bucket name, alternatively
                              S3_BUCKET env variable can also be used
      --cert=<certPath>     use this option to pass certificate path of the S3,
                              alternatively S3_CERT_PATH env variable can also
                              be used
      --content-type=<contentType>
                            optinoal flag to pass content type of the file used
                              for upload operation, for standard file like pdf,
                              txt, etc. appropriate content type will be set ,
                              alternatively use env var S3_CONTENT_TYPE
      --endpoint=<endpointUrl>
                            use this option to pass endpotint url,
                              alternatively S3_ENDPOINT env variable can also
                              be used
      --file=<file>         pass the single file to be uploaded to the bucket
                              only used for upload operation, alternatively
                              INPUT_FILE env variable can also be used
  -h, --help                display command usage info
      --operation=<operation>
                            supported operation options are list|create|upload
      --region=<s3region>   S3 region, defaults to us-west-1, alternatively
                              S3_REGION env variable can also be used
      --secret-key=<secretKey>
                            use this option to pass secret key, alternatively
                              S3_SECRET_KEY env variable can also be used
```
<img width="2044" height="1248" alt="image" src="https://github.com/user-attachments/assets/d4bce3ee-c811-4922-b98e-e976b629f341" />


Instead of passing the keys in CLI argument these can be set in environment variables. In Gitbash or WSL2 we can use export command to configure environment values to variable for the shell. Sample command like below where the keys are fetched from the seaweedfs secrets.

```
export S3_ACCESS_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "accessKey" | base64decode}}')
export S3_SECRET_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "secretKey" | base64decode}}')
```


For successful execution see below example

Successful exection Command
```
$ jbang "app.java" --operation list --endpoint https://s3.swfs.com --cert seaweed-s3/cert.pem
```

Output of CLI using  list operation to list buckets

<img width="2286" height="294" alt="image" src="https://github.com/user-attachments/assets/00d4baff-0497-46b8-bcfe-5245d79580b3" />

Output of CLI with create operation to create new buckets 

<img width="2478" height="290" alt="image" src="https://github.com/user-attachments/assets/e0db34f3-d7b6-4001-9154-f7ef98cfbf5a" />

Output of CLI with upload operation to upload single file

<img width="2476" height="364" alt="image" src="https://github.com/user-attachments/assets/216d93fe-267c-4c86-9425-a27a94db75c8" />

List of buckets after creation of bucket
<img width="2408" height="296" alt="image" src="https://github.com/user-attachments/assets/c4de7971-1371-4a30-835a-06e0e73a389f" />


To generate the PEM certificate format use the attached script with specific dns, like in below example

```
./getCertificate.sh --dns s3.swfs.com --port 443
```
Output
<img width="1434" height="428" alt="image" src="https://github.com/user-attachments/assets/e97472c5-4078-4169-bfdd-68e17fb024ce" />
