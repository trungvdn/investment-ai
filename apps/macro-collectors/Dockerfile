FROM python:3.12-slim
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends poppler-utils ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && useradd -u 1000 -m collector && mkdir /collector-data && chown collector:collector /collector-data
WORKDIR /app
COPY macro_collectors ./macro_collectors
ENV PYTHONUNBUFFERED=1
USER collector
ENTRYPOINT ["python", "-m", "macro_collectors"]
