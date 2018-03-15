FROM alpine:latest

COPY httpmq /bin/httpmq

EXPOSE 1218
CMD [ "httpmq" ]