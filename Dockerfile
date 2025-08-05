FROM gcr.io/distroless/static-debian12
ARG BIN_NAME
ADD ./bin/dist/${BIN_NAME} /
COPY ./config.yaml /
COPY ./config.test.yaml /

EXPOSE 80
EXPOSE 50051