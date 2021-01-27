FROM python:2-slim-stretch

LABEL maintainer="rpzhang@xxxxx.cn"

# Workdir
WORKDIR /platform/auth

# Env prepare
COPY sdk/python/requirements.txt .
RUN pip install -r requirements.txt

COPY sdk/python .

CMD ["python", "example/main.py"]
