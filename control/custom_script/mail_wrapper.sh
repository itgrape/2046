#!/bin/bash

JOB_ID=$SLURM_JOB_ID
JOB_NAME=$SLURM_JOB_NAME
JOB_USER=$SLURM_JOB_USER
JOB_STATE=$SLURM_JOB_STATE
MAIL_TYPE=$SLURM_JOB_MAIL_TYPE
MAIL_USER=$(scontrol show job $JOB_ID | grep -oP '(?<=MailUser=)[^ ]*')


SUBJECT="$JOB_ID - $JOB_NAME - $MAIL_TYPE"
BODY="用户名称: $JOB_USER\n作业ID: $JOB_ID\n作业名称: $JOB_NAME\n作业状态: $JOB_STATE\n\n"

# echo -e "To: $MAIL_USER\nSubject: $SUBJECT\n\n$BODY" | /usr/sbin/sendmail -v -t >> /var/log/slurm/sendmail.log

curl -X POST http://api.pushihao.com/v1/email -H "Content-Type: application/json" -d << EOF
{
    "secretKey":"pushihao",
    "fromName": "Slurm",
    "fromEmail": "imag@njust.edu.cn",
    "subject": "$SUBJECT",
    "htmlContent": "$BODY",
    "toName": "Dear $JOB_USER",
    "toEmail": "$MAIL_USER"
}
EOF