## DisconnectEnvironment action

Disconnect from an environment. This action will disconnect open websockets towards sense and events. The action is not needed for most scenarios, however if a scenario mixes different types of environmentsor uses custom actions towards external environment, it should be used directly after the last action towards the environment.

Since the action also disconnects any open websocket to Sense apps, it does not need to be preceeded with a `disconnectapp` action.
