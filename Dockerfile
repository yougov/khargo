FROM scratch
  
WORKDIR /

COPY khargo /

EXPOSE 8000

ENTRYPOINT ["/khargo"]