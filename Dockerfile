FROM alpine
COPY ./bin/linux_amd64/maoni ./bin/maoni
RUN chmod +x /bin/maoni
CMD ["/bin/maoni"]