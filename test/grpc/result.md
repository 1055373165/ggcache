INFO[0009] [Server localhost:9999] Received RPC request - group: scores, key: Hannah Harris 
INFO[0009] key Hannah Harris not found in cache, loading locally 
INFO[0009] current server localhost:9999, key Hannah Harris  
DEBU[0009] key Hannah Harris is mapped to remote peer localhost:10001 
INFO[0009] fetching scores/Hannah Harris from peer localhost:10001 reusing connection 

INFO[0002] [Server localhost:10001] Received RPC request - group: scores, key: Hannah Harris 
INFO[0002] key Hannah Harris not found in cache, loading locally 
INFO[0002] current server localhost:10001, key Hannah Harris  
DEBU[0002] key Hannah Harris is mapped to current node localhost:10001 (local handling) 
INFO[0002] key Hannah Harris picked to local            
INFO[0002] Successfully retrieved student Hannah Harris score: 9043.00 
DEBU[0002] Database query time: 0 ms                    
INFO[0002] Update to cache: key=Hannah Harris, value=9043.00 
INFO[0002] Successfully retrieved student Hannah Harris score: 9043.00 locally 
DEBU[0009] the duration of this grpc Call is: 0 ms      
INFO[0009] Successfully retrieved student Hannah Harris score: 9043.00 from remote peer 
INFO[0009] populate cache key: Hannah Harris, value: 9043.00 in current node

INFO[0004] [Server localhost:10001] Received RPC request - group: scores, key: Hannah Harris 
INFO[0004] key Hannah Harris found in cache, return directly 