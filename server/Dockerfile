FROM python:3.9.7-slim

RUN apt-get update && apt-get install -y netcat-openbsd && rm -rf /var/lib/apt/lists/*

COPY server /
RUN python -m unittest tests/test_common.py
ENTRYPOINT ["/bin/sh"]
