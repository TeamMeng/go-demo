local token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IlRlYW1BbGljZUAxNjMuY29tIiwiZXhwIjoxNzc1NzgxODE4LCJ1c2VyX2FnZW50IjoidnNjb2RlLXJlc3RjbGllbnQiLCJ1c2VyX2lkIjoiYzYzY2ZiN2EtZWE1OS00ODk4LTg5OTUtNmEzOWYwNjkxNmZiIn0.4LVe0YK_zejLAl5A3XtPdIYvJpB5RHrhR-pvTEX9YO8"
local path = "/todos/2"

function init(args)
  if #args >= 1 and args[1] ~= "" then
    if args[1]:sub(1, 1) == "/" then
      path = args[1]
    else
      path = "/todos/" .. args[1]
    end
  end

  if #args >= 2 and args[2] ~= "" then
    token = args[2]
  end

  wrk.method = "GET"
  wrk.headers["Accept"] = "application/json"
  wrk.headers["Authorization"] = "Bearer " .. token
  wrk.headers["User-Agent"] = "vscode-restclient"
end

function request()
  wrk.path = path
  return wrk.format(nil, path)
end

function response(status, headers, body)
  if status ~= 200 and status ~= 401 and status ~= 404 and status ~= 429 then
    io.stderr:write(string.format("unexpected status=%d body=%s\n", status, body))
  end
end
