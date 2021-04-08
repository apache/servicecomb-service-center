Heartbeat
========================
Heartbeat configuration. Configure app.yaml according to your needs.

::

   heartbeat:
     # configuration of websocket long connection
     websocket:
       pingInterval: 30s
     # heartbeat.kind="checker or cache"
     # if heartbeat.kind equals to 'cache', should set cacheCapacity,workerNum and taskTimeout
     # capacity = 10000
     # workerNum = 10
     # timeout = 10
     kind: cache
     cacheCapacity: 10000
     workerNum: 10
     timeout: 10

.. list-table::
  :widths: 15 20 5 10
  :header-rows: 1

  * - field
    - description
    - required
    - value
  * - heartbeat.websocket.pingInterval
    - websocket ping interval.
    - yes
    - like 30s
  * - heartbeat.kind
    - there are two types of heartbeat plug-ins. With cache and without cache.
    - yes
    - cache/checker
  * - heartbeat.cacheCapacity
    - cache capacity
    - yes
    - a integer, like 10000
  * - heartbeat.workerNum
    - the number of working cooperations
    - yes
    - a integer, like 10
  * - heartbeat.timeout
    - processing task timeout (default unit: s)
    - yes
    - a integer, like 10