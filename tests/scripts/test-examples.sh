for CONFIG in $(find ./examples/ -type f)
do
    echo $CONFIG
    ./bin/kafka-traffic-generator --config ${CONFIG}
done
