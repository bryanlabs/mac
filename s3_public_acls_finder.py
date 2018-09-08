import boto3 
import requests
import json

s3_client_connection = boto3.resource(
    's3'
)

def check_bucket_grant(grant_permission, bucket_name):
    granted_warning = 'The following permission: *{}* has been granted on the bucket *{}*'
    if grant_permission == 'read':
        return granted_warning.format('Read - Public Access: List Objects', bucket_name), True
    elif grant_permission == 'write':
        return granted_warning.format('Write - Public Access: Write Objects', bucket_name), True
    elif grant_permission == 'read_acp':
        return granted_warning.format('Write - Public Access: Read Bucket Permissions', bucket_name), True
    elif grant_permission == 'write_acp':
        return granted_warning.format('Write - Public Access: Write Bucket Permissions', bucket_name), True
    elif grant_permission == 'full_control':
        return granted_warning.format('Public Access: Full Control', bucket_name), True
    return ''

def check_S3_buckets_grants():
    for bucket in s3_client_connection.buckets.all():
        # print(bucket.name)
        acl = bucket.Acl()
        for grant in acl.grants:
            #http://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html
            if grant['Grantee']['Type'].lower() == 'group' \
                and grant['Grantee']['URI'] == 'http://acs.amazonaws.com/groups/global/AllUsers':
           
                text_to_post = check_bucket_grant(grant['Permission'].lower(), bucket.name)
                print(text_to_post)

if __name__ == "__main__":
    check_S3_buckets_grants()