### Simple CLI built using Jbang to access the S3 compatible service with some basic operations

Pre-requistes
  - S3 compatible store is deployed, for development in my case used KinD deployed with Seaweedfs and to expose the S3 gateway as SSL, cert-manager and Apisix route were used.

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
> jbang "s3cliapp\app.java" --operation list --cert ./cert.pem
```

Output

```
Missing required options: '--endpoint=<endpointUrl>', '--access-key=<accessKey>', '--secret-key=<secretKey>'
Usage: s3cli [-h] --access-key=<accessKey> [--bucket=<bucketName>]
             [--cert=<certPath>] [--content-type=<contentType>]
             --endpoint=<endpointUrl> [--file=<file>] --operation=<operation>
             [--region=<s3region>] --secret-key=<secretKey>
s3cli operations create and list buckets, upload file.
      --access-key=<accessKey>

      --bucket=<bucketName> bucket name
      --cert=<certPath>     certificate path of the S3
      --content-type=<contentType>
                            file to upload when using upload operation
      --endpoint=<endpointUrl>

      --file=<file>         file to upload when using upload operation
  -h, --help                display help message
      --operation=<operation>
                            operation list|create|upload
      --region=<s3region>
      --secret-key=<secretKey>
```

Instead of passing the keys in CLI argument these can be set in environment variables. In Gitbash or WSL2 we can use export command to configure environment values to variable for the shell. Sample command like below where the keys are fetched from the seaweedfs secrets.

```
export S3_ACCESS_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "accessKey" | base64decode}}')
export S3_SECRET_KEY=$(kubectl get -n seaweedfs secret admin-s3-secret -o go-template='{{index .data "secretKey" | base64decode}}')
```


For successful execution see below example

```
$ jbang "app.java" --operation list --endpoint https://s3.swfs.com --cert /c/thiru/edu/gitsource/Learnings/seaweed-s3/cert.pem
[jbang] Building jar for app.java...
Input provided for S3 service [operation='list', endpoint='https://s3.swfs.com', accessKey='29GMH9D238BZR5P5MH3JJ', secretKey='*****', certPath='C:/thiru/edu/gitsource/Learnings/seaweed-s3/cert.pem', region='us-west-1', bucketName='null', file='null', contentType='']
------------------------
No buckets found
------------------------
Execution Completed!!!
```

To generate the PEM certificate format use the attached script with specific dns, like in below example

```
./getCertificate.sh --dns s3.swfs.com --port 443
```