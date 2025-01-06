# System Architecture

## High-Level Architecture Diagram

<div class="mermaid">
%%{init: {'theme': 'forest', 'themeVariables': {'primaryColor': '#00ffcc', 'edgeLabelBackground': '#fefefe'}}}%%
flowchart TB
    subgraph LocalMachine[Local Machine]
        direction TB

        subgraph UserSpace[User Space]
            style UserSpace stroke-dasharray: 5 5
            dApp1[dApp 1]
            dApp2[dApp 2]
        end

        subgraph Daemons[daemons]
            style Daemons stroke-dasharray: 5 5
            Khedra[Khedra]
            Geth[Geth / Erigon / Reth]
        end

        dApp1 <--> |rpc| Geth
        dApp2 <--> |rpc| Geth
        Geth <--> |rpc| Khedra
        Khedra <--> |sdk| dApp1
        Khedra <--> |sdk| dApp2
    end

    Geth <--> |peer-2-peer| Internet["Internet / 'The Chain'"]
</div>
