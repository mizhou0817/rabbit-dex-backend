Rabbit dex backend - hybrid exchange with starknet settlement layer.

**Code:**
1. model - main golang data types
2. model/tnt - tarantool cartridge app
3. model/tnt/app - main app logic on lua
4. tests - risk tests for engine
5. configs-example - example of the configs required by services. (should be copied to ~/.rabbit)
6. each other golang service inside its own folder

**How to install:**

**_Step0 - Install tarantool and cartridge_**

```shell
Required version 2.10.2+  - 2.10.0 has a serious performace bug.
https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/installation/ #follow this instruction for installing tarantool and cartridge

tarantool --version # should be 2.10.2+
cartridge --version # should be 2.12.2+
```


**_Step1 - build tarantool app_**

```shell
cd model/tnt        #you should be in ./model/tnt folder
cartridge build     # this will build the cartridge app
./run-reset.sh     # this will start and configure the cluster
```

**_Step2 - copy defulat configs_**

```shell
mkdir ~/.rabbit
cp -rf ./configs-example/* ~/.rabbit/
```

**_Step3 - install centrifugo service, required json version_**

```shell
https://github.com/centrifugal/centrifugo/tree/tarantool_use_json  # clone and build this version of centrifugo 
```

**_Step4 - check that tests passed_**
```shell
cd tests
make all_sequences
```
