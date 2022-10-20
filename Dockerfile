FROM gcr.io/distroless/static-debian11:nonroot

ENTRYPOINT ["/usr/bin/vattenfall"]
COPY vattenfall /usr/bin/vattenfall
