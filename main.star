# This Kurtosis package spins up a minimal GM rollup that connects to a DA node
#
# NOTE: currently this is only connecting to a local DA node

da_node = import_module("github.com/rollkit/local-da/main.star@v0.3.0")

def run(plan):
    ##########
    # DA
    ##########

    da_address = da_node.run(
        plan,
    )
    plan.print("connecting to da layer via {0}".format(da_address))

    #####
    # Artela-rollkit
    #####

    plan.print("Adding Artela-rollkit service")
    plan.print("NOTE: This can take a few minutes to start up...")
    artroll_start_cmd = [
        "rollkit",
        "start",
        "--rollkit.aggregator",
        "--rollkit.da_address {0}".format(da_address),
    ]
    artroll_port_number = 26657
    artroll_port_spec = PortSpec(
        number=artroll_port_number, transport_protocol="TCP", application_protocol="http"
    )
    artroll_frontend_port_spec = PortSpec(
        number=1317, transport_protocol="TCP", application_protocol="http"
    )
    artroll_ports = {
        "jsonrpc": artroll_port_spec,
        "frontend": artroll_frontend_port_spec,
    }
    artroll = plan.add_service(
        name="artroll",
        config=ServiceConfig(
            # Using rollkit version v0.13.5
            image="ghcr.io/rollkit/gm:05bd40e",
            cmd=["/bin/sh", "-c", " ".join(gm_start_cmd)],
            ports=gm_ports,
            public_ports=gm_ports,
            ready_conditions=ReadyCondition(
                recipe=ExecRecipe(
                    command=["rollkit", "status"],
                    extract={
                        "output": "fromjson | .node_info.network",
                    },
                ),
                field="extract.output",
                assertion="==",
                target_value="gm",
                interval="1s",
                timeout="1m",
            ),
        ),
    )
