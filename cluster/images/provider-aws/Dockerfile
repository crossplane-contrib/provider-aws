FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TINI_VERSION

ADD provider /usr/local/bin/crossplane-aws-provider

COPY package /

EXPOSE 8080
USER 1001
ENTRYPOINT ["crossplane-aws-provider"]
