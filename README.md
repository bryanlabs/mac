# mac
**Multi Account Command** - runs any command in the context of the AWS Profile(s) specified.


You want to quickly interact with multiple AWS accounts without the complications of swapping access keys / environment variables... **mac** can help.

**How it works:**  
Each account you manage has a mac-service role that can be assumed via sts.
You have a mac-admins group on your source account with permissions to assume role into the various accounts with the mac-service role.
You have a mac-accounts policy attached to the group stating which accounts they can assume role into.
When a new admin comes on board, you create them an IAM user with access keys, then add them to the mac-admins group.
When a new account is created you run the cloud formation template to install the service role, and then update the mac-accounts policy on the source account.
The admin prefixes any normal command, with the account(s) of interest. The command is executed in the context of each account, and the results are displayed via stdout.

**USAGE:** 

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


**SETUP:**

**Cross Accounts:** Deploy the mac-service cloud formation template in all accounts that you wish to admin including any source accounts.  
<span style="color:red">**NOTE:** </span> Update "SOURCEACCOUNT" to the account number for your Source Account.

````
{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "Creates an IAM role with cross-account access for mac.",
    "Metadata": {
        "VersionDate": {
            "Value": "20170717"
        },
        "Identifier": {
            "Value": "mac-service"
        }
    },
    "Resources": {
        "CrossAccountRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "RoleName": "mac-service",
                "AssumeRolePolicyDocument": {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "AWS": [
                                    "arn:aws:iam::SOURCEACCOUNT:root"
                                ]
                            },
                            "Action": "sts:AssumeRole"
                        }
                    ]
                },
                "Path": "/",
                "Policies": [
                    {
                        "PolicyName": "AdministratorAccess",
                        "PolicyDocument": {
                            "Version": "2012-10-17",
                            "Statement": [
                                {
                                    "Effect": "Allow",
                                    "Action": "*",
                                    "Resource": "*"
                                }
                            ]
                        }
                    }
                ]
            }
        }
    }
}
````


**Source Account:**  

Create an IAM group called "mac-admins"  
Create an IAM Policy called "mac-accounts"  
<span style="color:red">**NOTE:** </span> each account you admin should be added to this list.
  

````
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": [
        "arn:aws:iam::111111111111:role/mac-service",
        "arn:aws:iam::222222222222:role/mac-service"
      ]
    }
  ]
}
````

create role mac_service  
attach policy_mac_accounts to IAM_group mac_admins  
attach policy Admin Access to mac_service  
add IAM users to mac_admins  


**Administrator environment setup:**

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