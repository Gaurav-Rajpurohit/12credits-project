-- File: wrk_downtime.lua
-- Logs response failures and computes overall downtime during a deployment swap.

local failed_requests = 0
local total_requests = 0

-- Called for every response received
response = function(status, headers, body)
   total_requests = total_requests + 1
   if status < 200 or status >= 400 then
      failed_requests = failed_requests + 1
      io.write(string.format("[FAIL] Status Code: %d | Time Offset: %.3f sec\n", status, os.clock()))
   end
end

-- Called at the end of the benchmark
done = function(summary, latency, requests)
   print("\n==================================================")
   print("           DOWNTIME PROFILE RESULTS               ")
   print("==================================================")
   print(string.format("Total Transmissions:   %d", total_requests))
   print(string.format("Successful Requests:   %d", total_requests - failed_requests))
   print(string.format("Failed Connections:    %d", failed_requests))
   
   local success_rate = 100.0
   if total_requests > 0 then
      success_rate = ((total_requests - failed_requests) / total_requests) * 100
   end
   print(string.format("Success Percentage:    %.4f%%", success_rate))
   
   -- Duration is in microseconds
   local duration_sec = summary.duration / 1000000.0
   local avg_throughput = summary.requests / duration_sec
   print(string.format("Avg Throughput (TPS):  %.2f req/sec", avg_throughput))
   
   -- Downtime = Failed / Throughput
   local estimated_downtime_ms = 0.0
   if avg_throughput > 0 then
      estimated_downtime_ms = (failed_requests / avg_throughput) * 1000.0
   end
   
   print(string.format("Measured Downtime:     %.2f ms", estimated_downtime_ms))
   print("==================================================")
end
