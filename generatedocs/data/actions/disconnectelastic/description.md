## DisconnectElastic action

Disconnect from a QSEoK environment. This action will disconnect open websockets towards sense and events. The action is not needed for most scenarios, however if a scenario mixes "elastic" environments with QSEoW or uses custom actions towards other type of environments, it should be used directly after the last action towards the elastic environment.

Since the action also disconnects any open websocket to Sense apps, and does not need to be preceeded with a `disconnectapp` action.
