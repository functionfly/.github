FROM debian:trixie-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    python3.12 \
    python3.12-venv \
    python3-pip \
    curl \
    jq \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sSf https://bootstrap.pypa.io/get-pip.py | python3.12 || true

COPY bootstrap.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/bootstrap.sh

ENV PATH="/usr/local/bin:${PATH}"
ENV PYTHONUNBUFFERED=1

ENTRYPOINT ["/usr/local/bin/bootstrap.sh"]
