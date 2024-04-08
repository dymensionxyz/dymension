############################################################################################################


rollex_1233-1


## 1. Trigger hub genesis event

Find the channel id:

    rollapp-evm q ibc channel end transfer channel-0

    dymd tx rollapp genesis-event $r channel-1 --from local-user --keyring-backend test --broadcast-mode block --gas auto -y --fees 1dym --chain-id dymension_100-1

Hub genesis_state:

    dymd q rollapp show $r
    dymd q bank denom-metadata | grep alxx -A 8

Balances after:

    rollapp-evm q bank balances $(rollapp-evm keys show rol-user --keyring-backend test -a)
    dymd q bank balances $(dymd keys show bob-genesis --keyring-backend test -a)
    dymd q bank balances $(dymd keys show alice-genesis --keyring-backend test -a)
    rollapp-evm q bank balances $(rollapp-evm q ibc-transfer escrow-address "transfer" channel-0)

## 2. Trigger rollapp genesis event

    rollapp-evm tx hubgenesis genesis-event dymension_100-1 channel-0 --from rol-user --keyring-backend test --broadcast-mode block -y

evm escrow: 60000000000000000000000 alxx (bob-genesis + alice-genesis)

### 3. Transfer hub -> hub:

send 5000 $d from hub bob-genesis to hub alice-genesis:

    dymd tx bank send bob-genesis --keyring-backend test $(dymd keys show alice-genesis --keyring-backend test -a) 5000$d --broadcast-mode block -y

Balances after:

hub bob-genesis: 9999999999999999995000 $d
hub alice-genesis: 50000000000000000005000 $d

### 4. IBC Transfer hub -> rollapp:

Create new account:

alex: ethm186s60vuy20qvpacxcezrg0efn5e7st5ag3seak

send 2500 ibc/9A5A1A1E5E45949DFC21299A9472B46626E36B21BB2F971105E8DF708573A72B from hub bob-genesis to evm alex:

    dymd tx ibc-transfer transfer transfer channel-60 $(rollapp-evm keys show alex --keyring-backend test -a) 2500$d --from bob-genesis --keyring-backend test --broadcast-mode block -y

Balances after:

    rollapp-evm q bank balances $(rollapp-evm keys show alex --keyring-backend test -a)
    dymd q bank balances $(dymd keys show bob-genesis --keyring-backend test -a)
    dymd q bank balances $(dymd keys show alice-genesis --keyring-backend test -a)
    rollapp-evm q bank balances $(rollapp-evm q ibc-transfer escrow-address "transfer" channel-0)

Balances after:
  evm alex: 500 alxx
  hub bob-genesis: 9999999999999999994250 $d
  hub alice-genesis: 50000000000000000005000 $d
  evm escrow: 59999999999999999999500 alxx

### IBC Transfer rollapp -> hub:

Balances before
  evm alex: 216 alxx
  hub bob-genesis: 9999999999999999994000 $d
  hub alice-genesis: 50000000000000000005333 $d
  evm escrow: 59999999999999999999784 alxx

send 21 alxx from evm alex to hub alice-genesis:

    rollapp-evm tx ibc-transfer transfer transfer channel-0 $(dymd keys show alice-genesis --keyring-backend test -a) 21alxx --from alex --keyring-backend test --broadcast-mode block -y

Balances after
  evm alex: 195 alxx
  hub bob-genesis: 9999999999999999994000 $d
  hub alice-genesis: 50000000000000000005333 $d
  evm escrow: 59999999999999999999805 alxx

Relayer error: `Apr 07 00:01:26 ip-172-31-16-109 rly[38702]: 2024-04-07T00:01:26.946347Z       
warn        Flush not complete       
{"error": "failed to enqueue pending messages for flush: no ibc messages found for write_acknowledgement query: 
write_acknowledgement.packet_dst_channel='channel-60' AND write_acknowledgement.packet_sequence='1'"}`

Apr 07 01:06:19 ip-172-31-16-109 rly[38702]: 2024-04-07T01:06:19.555019Z        info        Successful transaction        {"provider_type": "cosmos", "chain_id": "rollex_1234-1", "packet_src_channel": "channel-0", "packet_dst_channel": "channel-60", "gas_used": 163602, "fees": "", "fee_payer": "ethm1fwjthczkd9k46yxvgc8qehy2ayms8zr6lwkw4m", "height": 1756, "msg_types": ["/ibc.core.client.v1.MsgUpdateClient", "/ibc.core.channel.v1.MsgAcknowledgement"], "tx_hash": "A9CD449102B42FAEF551EF5490B799C926159CC4160BCF795B21CD15DC40D170"}


867800 - 866969 = 831
831 * 4,75 / 60 = 65,7875 minutes

02:00 - 03:06 = 66 minutes


### eIBC Transfer rollapp -> hub:

send 2000 alxx from evm alex to hub alice-genesis:

    rollapp-evm tx ibc-transfer transfer transfer channel-0 $(dymd keys show alice-genesis --keyring-backend test -a) 2000alxx --from alex --keyring-backend test --broadcast-mode block -y --memo '{"eibc": {"fee": "350"}}'

Balances after:
  evm alex: 145 alxx
  hub bob-genesis: 9999999999999999994000 $d
  hub alice-genesis: 50000000000000000005333 $d
  evm escrow: 59999999999999999999855 alxx

check demand orders:

    dymd q eibc list-demand-orders PENDING

bob fulfills order:

    dymd tx eibc fulfill-order d03fb317233c641065d6b1024920117f2d4a78bf950b0c3bb2e8edb511ffad1f --from bob-genesis --keyring-backend test --broadcast-mode block -y

demand_orders:
- fee:
  - amount: "20"
    denom: $d
    id: 43179f65b72f3b213457e0c5a16f998ca3ed018f8d8010ba62753027858b63bb
    is_fullfilled: true
    price:
  - amount: "30"
    denom: $d
    recipient: dym1g3w9nvkg70h0mhhmqtm2wutzvfnc7gwzyyf5xt
    tracking_packet_key: "\0\x01/rollex_1234-1/\0\0\0\0\0\0\ts/channel-60\0\0\0\0\0\0\0\x06"
    tracking_packet_status: PENDING


Balances after:
  evm alex: 145 alxx
  hub bob-genesis: 9999999999999999993970 $d
  hub alice-genesis: 50000000000000000005363 $d
  evm escrow: 59999999999999999999855 alxx

Balances after dispute period:
  evm alex: 145 alxx
  hub bob-genesis: 9999999999999999993970 $d
  hub alice-genesis: 50000000000000000005363 $d
  evm escrow: 59999999999999999999855 alxx


    rollapp-evm tx ibc-transfer transfer transfer channel-0 $(dymd keys show bob-genesis --keyring-backend test -a) 60alxx --from alex --keyring-backend test --broadcast-mode block -y --memo '{"eibc": {"fee": "10"}}'
    dymd tx eibc fulfill-order 7a7824c8c4015d257b4115b79b4806db1164b41f1317039f1f2bf693c109badf --from bob-genesis --keyring-backend test --broadcast-mode block -y


Balances after dispute period:
  evm alex: 85 alxx
  hub bob-genesis: 9999999999999999993970 $d
  hub alice-genesis: 50000000000000000005363 $d
  evm escrow: 59999999999999999999915 alxx

### eIBC Transfer hub -> rollapp with timeout:

Balances before
  evm alex: 2000 alxx
  hub bob-genesis: 9999999999999999996350 $d
  hub alice-genesis: 50000000000000000001650 $d
  evm escrow: 59999999999999999998000 alxx

send 133330000 $d from hub bob-genesis to evm alex:

    dymd tx ibc-transfer transfer transfer channel-60 $(rollapp-evm keys show alex --keyring-backend test -a) 133330000$d --from bob-genesis --keyring-backend test --broadcast-mode block -y --packet-timeout-timestamp 0 --packet-timeout-height "1-$(($(rollapp-evm status | jq -r '.SyncInfo.latest_block_height') ))"


    rollapp-evm q bank balances $(rollapp-evm keys show alex --keyring-backend test -a)
    dymd q bank balances $(dymd keys show bob-genesis --keyring-backend test -a)


5000
9999999999999999993020

9999999999999866663020
133330000
9999999999999999993020

    ibc/61F462A67A1B95B80BBCC53696D322EC98621EBA3412A15AD79EC134507EC601

    dymd tx ibc-transfer transfer transfer channel-1 $(rollapp-evm keys show alex --keyring-backend test -a) 34001$d --from bob-genesis --keyring-backend test --broadcast-mode block -y --fees 1dym --packet-timeout-timestamp 1
9999999999999999931998
9999999999999999897997
    rollapp-evm tx bank send rol-user --keyring-backend test 1000000000000alxx ethm1pqnyjn867szetu0v2cc8zftf27wh83hrgdyjk9 -b block -y

Balances after
  evm alex: 500 alxx
  hub bob-genesis: 9999999999999999997850 $d
  hub alice-genesis: 50000000000000000001650 $d
  evm escrow: 59999999999999999999500 alxx

    rollapp-evm tx ibc-transfer transfer transfer channel-0 $(dymd keys show bob-genesis --keyring-backend test -a) 330alxx --from alex --keyring-backend test --broadcast-mode block -y --memo '{"eibc": {"fee": "321"}}'
   
    dymd q eibc list-demand-orders PENDING
    dymd tx eibc fulfill-order fc008a8cafcc77e5e479faebf738d5419bcc7622ea9f317a1f130bbf6ec090cb --from local-user --keyring-backend test --broadcast-mode block --fees 1dym -y


    
    export HUB_CHAIN_ID=dymension_100-1
    export HUB_RPC_URL=http://127.0.0.1:36657
    export ROLLAPP_CHAIN_ID=rollex_1440-1

