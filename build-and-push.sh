$(aws ecr get-login --no-include-email --region ap-northeast-1)
docker build -t terra/faucet .
docker tag terra/faucet:latest 616478479272.dkr.ecr.ap-northeast-1.amazonaws.com/terra/faucet:latest
docker push 616478479272.dkr.ecr.ap-northeast-1.amazonaws.com/terra/faucet:latest
