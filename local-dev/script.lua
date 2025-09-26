function transform(tag, ts, record)
  -- example: notice that a record has passed through a Lua filter
  record["lua_enriched"] = "true"
  -- example: duplicate the logging level in a convenient field (if available)
  if record["severity"] ~= nil and record["severity_text"] == nil then
    record["severity_text"] = tostring(record["severity"])
  end
  return 1, tag, ts, record
end
