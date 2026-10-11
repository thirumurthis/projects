package s3cliapp.support;

/*
 * Class uses PicoCLI library to pass option for the cli command
 */
import java.io.File;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.util.List;
import java.util.concurrent.Callable;

import org.springframework.stereotype.Component;

import picocli.CommandLine.Command;
import picocli.CommandLine.Option;
import picocli.CommandLine.Parameters;

@Component
@Command(name = "s3cli", mixinStandardHelpOptions = true, version = "s3cli 1.0",
        description = "s3cli operations create and list buckets, upload file.")
public class cliOptions implements Callable<Integer>{

    @Option(names={"--endpoint"}, defaultValue = "${env:S3_ENDPOINT}", required = true, description="use this option to pass endpotint url, alternatively S3_ENDPOINT env variable can also be used")
    private String endpointUrl;

    @Option(names={"--access-key"}, defaultValue = "${env:S3_ACCESS_KEY}", required = true, description="use this option to pass access key, alternatively S3_ACCESS_KEY env variable can also be used")
    private String accessKey;

    @Option(names={"--secret-key"}, defaultValue = "${env:S3_SECRET_KEY}", required = true, description="use this option to pass secret key, alternatively S3_SECRET_KEY env variable can also be used")
    private String secretKey;

    @Option(names={"--operation"}, defaultValue = "${env:S3_OPERATION}", required = true, description="supported operation options are list|create|upload")
    private String operation;

    @Option(names={"--region"}, defaultValue = "${env:S3_REGION:-us-west-1}", required = false, description="S3 region, defaults to us-west-1, alternatively S3_REGION env variable can also be used")
    private String s3region;

    @Option(names={"--bucket"}, defaultValue = "${env:S3_BUCKET}", required = false, description="use this option to pass bucket name, alternatively S3_BUCKET env variable can also be used")
    private String bucketName;

    @Option(names={"--cert"}, defaultValue = "${env:S3_CERT_PATH}", required = false,description="use this option to pass certificate path of the S3, alternatively S3_CERT_PATH env variable can also be used")
    private String certPath;

    @Option(names={"--file"}, defaultValue="${env:INPUT_FILE}", required = false, description="pass the single file to be uploaded to the bucket only used for upload operation, alternatively INPUT_FILE env variable can also be used")
    private String file;

    @Option(names={"--content-type"}, defaultValue="${env:S3_CONTENT_TYPE:-}", required = false, description="optinoal flag to pass content type of the file used for upload operation, for standard file like pdf, txt, etc. appropriate content type will be set , alternatively use env var S3_CONTENT_TYPE")
    private String contentType;

    @Option(names = {"-h", "--help"}, usageHelp = true, description = "display command usage info")
    private boolean helpflag = false;

    private final s3Handler s3clihandler; 

    public cliOptions(s3Handler s3clihandler){
        this.s3clihandler = s3clihandler;
    }

    @Override
    public Integer call() throws Exception{
        try{
            s3Object s3obj = new s3Object(operation, endpointUrl, accessKey, secretKey, certPath, s3region, bucketName, file, contentType);
            System.out.println(s3clihandler.handler(s3obj));
        } catch(Exception e){
            StringWriter sw = new StringWriter();
            PrintWriter pw = new PrintWriter(sw);
            //e.printStackTrace(pw);
            System.err.println(sw.toString());
            System.out.println(e.getMessage());
            return -1;
        }

        return 0;
    }
}
