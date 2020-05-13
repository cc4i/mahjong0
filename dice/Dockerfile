# Step 1 prepare dice
FROM golang:alpine as dicebuilder

ENV GO111MODULE=on
ARG ver=0.0.0

# Install git
RUN apk update && apk add git
COPY . $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/

#get dependancies
#you can also use dep
RUN go get -d -v

#build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/dice


##STEP 2 prepare awscli2
#FROM alpine:3.11 as awscli2installer
#
#ENV GLIBC_VER=2.31-r0
#
## install glibc compatibility for alpine
#RUN apk --no-cache add \
#        binutils \
#        curl \
#        bash \
#    && curl -sL https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
#    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-${GLIBC_VER}.apk \
#    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-bin-${GLIBC_VER}.apk \
#    && apk add --no-cache \
#        glibc-${GLIBC_VER}.apk \
#        glibc-bin-${GLIBC_VER}.apk \
#    && curl -sL https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o awscliv2.zip \
#    && unzip awscliv2.zip \
#    && aws/install \
#    && rm -rf \
#        awscliv2.zip \
#        aws \
#        /usr/local/aws-cli/v2/*/dist/aws_completer \
#        /usr/local/aws-cli/v2/*/dist/awscli/data/ac.index \
#        /usr/local/aws-cli/v2/*/dist/awscli/examples \
#    && rm glibc-${GLIBC_VER}.apk \
#    && rm glibc-bin-${GLIBC_VER}.apk \
#    && rm -rf /var/cache/apk/*


# STEP 3 build package
FROM alpine:3.11

ARG BUILD_DATE

LABEL org.label-schema.name="dice-all-in-one" \
      org.label-schema.git="https://github.com/cc4i/mahjong0.git" \
      org.label-schema.build-date=$BUILD_DATE


ENV KUBE_LATEST_VERSION="v1.16.9"
ENV AWS_IAM_AUTHENTICATOR="0.5.0"
ENV HELM_VERSION="v3.2.1"
ENV GLIBC_VER=2.31-r0

RUN apk add --no-cache ca-certificates \
        bash \
        git \
        openssh \
        curl \
        nodejs \
        nodejs-npm \
    && wget -q https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && wget -q https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_IAM_AUTHENTICATOR}/aws-iam-authenticator_${AWS_IAM_AUTHENTICATOR}_linux_amd64 -O  /usr/local/bin/aws-iam-authenticator \
    && chmod +x /usr/local/bin/aws-iam-authenticator \
    && wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm \
    && curl -sL https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-${GLIBC_VER}.apk \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-bin-${GLIBC_VER}.apk \
    && apk add --no-cache \
        glibc-${GLIBC_VER}.apk \
        glibc-bin-${GLIBC_VER}.apk \
    && curl -sL https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o awscliv2.zip \
    && unzip awscliv2.zip \
    && aws/install \
    && rm -rf \
        awscliv2.zip \
        aws \
        /usr/local/aws-cli/v2/*/dist/aws_completer \
        /usr/local/aws-cli/v2/*/dist/awscli/data/ac.index \
        /usr/local/aws-cli/v2/*/dist/awscli/examples \
    && npm i -g aws-cdk \
    && rm glibc-${GLIBC_VER}.apk \
    && rm glibc-bin-${GLIBC_VER}.apk \
    && rm -rf /var/cache/apk/* \
    && rm -rf /root/.config \
    && rm -rf /root/.npm
# Copy dice
COPY --from=dicebuilder /go/bin/dice /usr/local/bin/dice
#COPY --from=awscli2installer /usr/local/aws-cli/ /usr/local/aws-cli/
#COPY --from=awscli2installer /usr/local/bin/ /usr/local/bin/

WORKDIR /workspace

CMD bash