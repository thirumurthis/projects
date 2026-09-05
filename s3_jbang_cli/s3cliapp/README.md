### S3 CLI using Jbang to access the S3 service for basic operations

#### Pre-requistes
  - Jbang CLI installed. JBang allows to execute Java like script, refer the [JBang](https://www.jbang.dev) documentation for more details.
  - S3 compatible service accessible or deployed in local. In my case have deployed Seaweedfs in KinD cluster using operators chart, cert-manager and Apisix route used to expose the HTTPS endpoint with self signed certs. For more details to deploy Seaweedfs refer my blog at [Hashnode](https://thirumurthi.hashnode.dev/deploy-s3-compatible-seaweedfs-in-kind-cluster) or [Medium](https://medium.com/@thirumurthi.s/s3-compatible-seaweedfs-service-deployed-in-kind-cluster-50ad382aec6a?sharedUserId=thirumurthi.s).

#### Summary

The idea of this code is to perform basic operation on the S3 using the AWS S3 SDK dependencies. The structure is managed for code maintenance, all the java code can be placed in single file as well. 

The Picocli dependency is used for create command line type interface, where we can pass arguments using flags. Spring Boot is used here since when the Picocli strater dependency is added to class path the factory bean is automatically injected to the context. The AWS S3 sdk is used to create the client using the provided certificate. This CLI requires certificate to be passed.

The Picocli library provides annotation support where the values of the flag can be read from the environment variables as well. The `@option` annotation in `cliOptions.java` could see the default value using `${env:S3_ENDPOINT}`. This helps to set some of the credentials variable to be set in the environment variable.

The application.yaml is added to the structure, used to control the logging level details. The code uses System.out to print the info to console when the CLI is executed. 

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

Below is the command to execute the S3 CLI app code using JBang, the command will look like below. In this case we are not passing any flags so will display the CLI usage like in the below output section. 

```
jbang s3cliapp\app.java
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

As mentioned Picocli supports to read variable values from environment, we can set the values to environment. Below is example, where the kubectl command is used to extract the key values from secret from Seaweedfs deployed server and set the value to shell env variable. Below will work in Git Bash, WSL2 and Linux terminals.

```
export S3_ACCESS_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "accessKey" | base64decode}}')
export S3_SECRET_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "secretKey" | base64decode}}')
```

Below shows the command to list the buckets from the S3 service, since the keys are passed via environment variables, the operation will be successfully completed. Refer the output snapshot.

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


To generate the certificate in PEM format use `openssl` command or the attached script can generate by passing specific dns and port. Which is shown below.

```
./getCertificate.sh --dns s3.swfs.com --port 443
```
Output
<img width="1434" height="428" alt="image" src="https://github.com/user-attachments/assets/e97472c5-4078-4169-bfdd-68e17fb024ce" />

##### Source code 
The source code for the JBang based S3 App CLI - [s3_jbang_cli/s3cliapp](https://github.com/thirumurthis/projects/edit/main/s3_jbang_cli/s3cliapp/README.md)
