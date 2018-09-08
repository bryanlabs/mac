# mac
**Multi Account Command** - runs any command in the context of the AWS Named Profile(s) specified.


You want to quickly interact with one or many AWS accounts without the complications of swapping access keys or environment variables... **mac** can help.

------------
How is this different from awsclis --profile option
------------
* awsclis --profile allows passing 1 profile at a time.  
* awsclis --profile only processes aws commands.  
* mac will loop through all the profiles you specify.  
* mac will run any binary, awscli, or your own.  (EG: findPublicBuckets.exe)

------------
How it works
------------
Each account you manage has a mac-service role that can be assumed via sts. IAM users join the mac-admins group, and inherit permissions to assumerole as defined in the mac-assumerole policy. When a new admin comes on board, you create them an IAM user with access keys, then add them to the mac-admins group. When a new Managed account is created you run the cloud formation template to install the service role. Also, update the stack on the IAM Account with the new Managed Account number. The admin prefixes any normal command, with the profile(s) of interest. The command will spawn parallel shells for each profile with the custom environment context, run the command,  and return the results.

------------
Account Type Definitions
------------
The following Account types will be mentioned, here are the definitions.
* **IAM Account:** This account is where all your IAM users exist. It's typicaly used in your central or security account.  
* **Managed Account:** This account is any account, including central/security that you want to be managed by the IAM Account users.  

------------
EXAMPLES 
------------

**Example 1: See all buckets in multiple accounts.**

````
mac -p 'bryanlabs,bryanlabsdev' 'aws s3 ls'
Profile: bryanlabsdev
2018-08-15 21:36:32 cf-templates-aviic4ggd7jk-us-east-1
Profile: bryanlabs
2017-12-16 17:34:50 bryanlabs
````
**Example 2: See the bucket policy from a specific account.**
````
# 
mac -p 'bryanlabs' 'aws s3api get-bucket-policy --bucket bryanlabs'
{
    "Policy": "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"AWSCloudTrailAclCheck20150319\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"cloudtrail.amazonaws.com\"},\"Action\":\"s3:GetBucketAcl\",\"Resource\":\"arn:aws:s3:::bryanlabs\"},{\"Sid\":\"AWSCloudTrailWrite20150319\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"cloudtrail.amazonaws.com\"},\"Action\":\"s3:PutObject\",\"Resource\":\"arn:aws:s3:::bryanlabs/CloudTrail/AWSLogs/111111111111/*\",\"Condition\":{\"StringEquals\":{\"s3:x-amz-acl\":\"bucket-owner-full-control\"}}}]}"
}
````

<span style="color:red">**NOTE:** </span> redirection and pipes don't work yet, for now wrap in a script.

**Example 3: Wrap complicated commands in scripts.**

````
Example:
cat <<EOF > runme.sh
#!/bin/bash
aws ec2 describe-instances | jq -r .[][].Instances[] | jq -r .InstanceId
EOF
mac -p 'bryanlabs' './runme.sh'
````


------------
SETUP
------------

**Managed Accounts:** Deploy the ManagedAccount.template in all accounts that you wish to admin including any IAM accounts.  

````
aws cloudformation create-stack --stack-name mac --template-body file://ManagedAccount.template --parameters ParameterKey=KeyPairName,ParameterValue=TestKey ParameterKey=SubnetIDs,ParameterValue=SubnetID1\\,SubnetID2
````


**IAM Account:** Deploy the IAMAccount.template in the Account where your IAM users are defined. Typically your central or security account.   

````
aws cloudformation create-stack --stack-name mac --template-body file://IAMAccount.template --parameters ParameterKey=KeyPairName,ParameterValue=TestKey ParameterKey=SubnetIDs,ParameterValue=SubnetID1\\,SubnetID2
````

<span style="color:red">**NOTE**: </span> My knowledge of cloudformation only allows assuming role into 1 Managed account. Others can be adding my manually modifying the inline policy, or submitting a merge request with the necessary changes.

------------
Administrator environment setup
------------

Configure Credentials: (.aws/credentials)

````
[default]
aws_access_key_id=AAAAABBBBBBCCCCCCDDDDDD
aws_secret_access_key=aA/bbbb/cccccccccc/DdDdDD/eeeeeEEE/Ff
````

Configure Named Profiles (.aws/config)
````
[default]
region=us-east-1

[profile centralservices]
region=us-east-1
role_arn=arn:aws:iam::111111111111:role/mac-service
source_profile=default

[profile security]
region=us-east-1
role_arn=arn:aws:iam::222222222222:role/mac-service
source_profile=default

[profile logging]
region=us-east-1
role_arn=arn:aws:iam::333333333333:role/mac-service
source_profile=default
````
Test:
````
mac -a 'centralservices,security' 'aws sts get-caller-identity'

dan@devbox:mac$ ./mac -a 'centralservices,security' 'aws sts get-caller-identity'
Profile: centralservices
{
    "Account": "centralservices",
    "UserId": "AROAJP7QAQIUQT6XY72VG:botocore-session-1534452856",
    "Arn": "arn:aws:sts::111111111111:assumed-role/mac-service/botocore-session-1534452856"
}
Profile: security
{
    "UserId": "AROAIWNRU4MMXGREHEURS:botocore-session-1534452839",
    "Arn": "arn:aws:sts::222222222222:assumed-role/mac-service/botocore-session-1534452839",
    "Account": "security"
}
````
