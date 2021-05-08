/usr/bin/rpk \
  redpanda start \
    --smp '1' \
    --reserve-memory 0M \
    --overprovisioned \
    --node-id '0' \
    --kafka-addr PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:9092 \
    --advertise-kafka-addr PLAINTEXT://redpanda:9092,OUTSIDE://localhost:9092
