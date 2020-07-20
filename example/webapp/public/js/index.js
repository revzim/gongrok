var app = new Vue({
  el: "#app",
  data: {
    strings: {
      title: "INIT NGROK CLIENT TUNNELS",
      protocols: {
        http: 0,
        tcp: 1,
        tls: 2,
      },

      btns: {
        initClient: "INIT CLIENT TUNNEL",
        disconnectClient: "DISCONNECT CLIENT TUNNEL",
        currentTunnel: "VIEW CURRENT TUNNEL",
      },
    },
    bools: {
      clientcreated: false,
    },
    ngrok: {},
    ngrokmockdata: {
      client: {
        id: ""
      },
      tunnel: {
        remoteaddr: "",
      }
    },
    formdata: {
      port: 0,
      tunnelname: "",
      host: "",
      protocol: 0,
    },
    formdataconfig: {
      tunnelname: FDataVals("tunnelname", "Tunnel Name", "ex: 'MyTunnelName'"),
      port: FDataVals("port", "Port", "ex: '8080'"),
      host: FDataVals("host", "Local Server Address", "ex: 'localhost'"),
      protocol: FDataVals("protocol", "Protocol", "ex: '0'"),
    },
  },
  methods: {
    checkInput(key) {
      switch (key) {
        case "port" || "protocol":
          // IF VALUE ISN'T NUMBER RETURN
          if (isNaN(parseInt(this.formdata[key], 10))) {
            this.formdata[key] = 0;
            return false;
          }
          break;
        case "tunnelname" || "host":
          if (typeof this.formdata[key] !== "string") {
            this.formdata[key] = "";
            return false;
          }
          break;
      }
      return true;
    },
    initClient() {
      let data = this.checkValues();
      if (!data.data || !data.keys) {
        return;
      }
      let keys = data.keys;
      axios.post("/client/new", data.data)
        .then((resp) => {
          // console.log("resp:", resp);
          if (resp.data.error) {
            alert(`client creation error ${resp.data.error}`)
            this.bools.clientcreated = false;
            return;
          }
          if (resp.data.client) {
            let data = resp.data;
            let client = data.client;
            let tunnel = data.tunnel;
            this.ngrok.data = {
              client: {
                ngroklocaladdr: client.ngroklocaladdr,
                opts: client.options,
                tunnels: client.tunnels,
                id: client.id,
              },
              tunnel: tunnel,
            }
            if (client.id !== "") {
              this.bools.clientcreated = true;
            }
          }

        })
        .catch((err) => {
          console.log("err:", err);
        })
    },
    disconnectClient() {
      // c0cebc0c-8299-4110-83e3-c1ec305638dc
      var formData = new FormData();
      formData.append("clientid", this.ngrok.data.client.id);
      axios.post("/client/disconnect", formData)
        .then((resp) => {
          // console.log("resp:", resp);
          let data = resp.data;
          // console.log("data:", data);
          if (data.error) {
            alert(`Error disconnecting client: ${data.error}`);
            return;
          }
          if (data.code === 200 && data.status === "OK") {
            // SUCCESSFULLY SHUTDOWN NGROK TUNNEL
            console.log("CLIENT SUCCESSFULLY SHUTDOWN!");
            this.ngrok.data = this.ngrokmockdata;
            this.bools.clientcreated = false;
            this.resetFormData();
          }
        })
        .catch((err) => {
          console.log("err:", err);
        })
    },
    resetFormData() {
      this.formdata.port = 0
      this.formdata.tunnelname = ""
      this.formdata.host = ""
      this.formdata.protocol = 0
    },
    checkValues() {
      var formData = new FormData();
      // SET VALS

      let keys = Object.keys(this.formdata);
      for (let key of keys) {
        console.log(key, this.formdata[key]);
        if (!this.checkInput(key)) {
          return { data: null, keys: null }
        }
        formData.append(key, this.formdata[key]);
      }
      return { data: formData, keys: keys }
    },
  },
  mounted() {
    this.ngrok = {
      data: this.ngrokmockdata
    }
  },
})

function FDataVals(key, title, placeholder) {
  return {
    key: key,
    title: title,
    placeholder: placeholder,
  }
}
