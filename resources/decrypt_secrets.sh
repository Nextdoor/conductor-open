# Decrypts secrets. Should be sourced.

# Helper to decrypt KMS blobs
kms_decrypt () {
    set -e
    local ENCRYPTED=$1
    local BLOB_PATH=$(mktemp)
    echo $ENCRYPTED | base64 --decode > $BLOB_PATH
    aws kms decrypt --ciphertext-blob fileb://$BLOB_PATH --output text --query Plaintext | base64 --decode
    rm -f $BLOB_PATH
}

SECRETS="GITHUB_ADMIN_TOKEN GITHUB_AUTH_CLIENT_SECRET GITHUB_WEBHOOK_SECRET JENKINS_PASSWORD SLACK_TOKEN JIRA_API_TOKEN"
# Try to decrypt the various KMS blobs, if they're set.
touch /tmp/secrets
for SECRET in $SECRETS; do
    BLOB_NAME="${SECRET}_BLOB"
    if [[ -n "${!BLOB_NAME}" ]]; then
        (
            echo "Decoding ${BLOB_NAME}."
            DECRYPTED=$(kms_decrypt "${!BLOB_NAME}")
            echo "${SECRET}='${DECRYPTED}'" >> /tmp/secrets
        ) &
    fi
done

wait

set -a
source /tmp/secrets
set +a
rm -rf /tmp/secrets
