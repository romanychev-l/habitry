FROM mirror.gcr.io/library/python:3.12-slim-bullseye as compile-image
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip \
 && pip install --no-cache-dir -r requirements.txt

FROM mirror.gcr.io/library/python:3.12-slim-bullseye as run-image
COPY --from=compile-image /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
WORKDIR /app
COPY bot ./bot
CMD ["python", "-m", "bot"]