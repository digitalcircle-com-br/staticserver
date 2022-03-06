#FROM req.digitalcircle.com.br/base
FROM scratch
COPY ./staticserver /staticserver
WORKDIR /
ENV TLS ""
ENV ROOT ""
ENV ADDR ""
ENTRYPOINT [ "/staticserver" ]
