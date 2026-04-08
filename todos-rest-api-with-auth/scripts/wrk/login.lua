local path = "/auth/login"
local users = {
  { email = "user@example.com", password = "password123" }
}
local invalid_users = {}
local valid_ratio = 1
local next_user_index = 1
local next_invalid_user_index = 1

local function trim(value)
  return (value:gsub("^%s+", ""):gsub("%s+$", ""))
end

local function json_escape(value)
  value = value:gsub("\\", "\\\\")
  value = value:gsub("\"", "\\\"")
  value = value:gsub("\b", "\\b")
  value = value:gsub("\f", "\\f")
  value = value:gsub("\n", "\\n")
  value = value:gsub("\r", "\\r")
  value = value:gsub("\t", "\\t")
  return value
end

local function build_body(user_email, user_password)
  return string.format(
    "{\"email\":\"%s\",\"password\":\"%s\"}",
    json_escape(user_email),
    json_escape(user_password)
  )
end

local function parse_user_line(line, line_number)
  local content = trim(line)
  if content == "" or content:sub(1, 1) == "#" then
    return nil
  end

  local email, password = content:match("^([^,]+),(.+)$")
  if not email or not password then
    error(string.format(
      "invalid users file format at line %d, expected: email,password",
      line_number
    ))
  end

  return {
    email = trim(email),
    password = trim(password),
  }
end

local function load_users(file_path)
  local file = io.open(file_path, "r")
  if not file then
    error("failed to open users file: " .. file_path)
  end

  local loaded_users = {}
  local line_number = 0

  for line in file:lines() do
    line_number = line_number + 1
    local user = parse_user_line(line, line_number)
    if user then
      loaded_users[#loaded_users + 1] = user
    end
  end

  file:close()

  if #loaded_users == 0 then
    error("users file is empty: " .. file_path)
  end

  return loaded_users
end

local function is_users_file_arg(value)
  return value:sub(1, 1) == "@"
end

local function parse_ratio(value)
  local ratio = tonumber(value)
  if ratio == nil or ratio < 0 or ratio > 1 then
    error("valid user ratio must be a number between 0 and 1")
  end
  return ratio
end

local function has_invalid_users()
  return #invalid_users > 0 and valid_ratio < 1
end

local function configure_from_args(args)
  if #args == 0 then
    return
  end

  if #args >= 3 and is_users_file_arg(args[1]) and is_users_file_arg(args[2]) then
    users = load_users(args[1]:sub(2))
    invalid_users = load_users(args[2]:sub(2))
    valid_ratio = parse_ratio(args[3])

    if #args >= 4 and args[4] ~= "" then
      path = args[4]
    end
    return
  end

  if is_users_file_arg(args[1]) then
    users = load_users(args[1]:sub(2))
    if #args >= 2 and args[2] ~= "" then
      path = args[2]
    end
    return
  end

  local email = args[1]
  local password = args[2]

  if email ~= nil and email ~= "" and password ~= nil and password ~= "" then
    users = {
      { email = email, password = password }
    }
  end

  if #args >= 3 and args[3] ~= "" then
    path = args[3]
  end
end

local function next_user()
  local user = users[next_user_index]
  next_user_index = next_user_index + 1
  if next_user_index > #users then
    next_user_index = 1
  end
  return user
end

local function next_invalid_user()
  local user = invalid_users[next_invalid_user_index]
  next_invalid_user_index = next_invalid_user_index + 1
  if next_invalid_user_index > #invalid_users then
    next_invalid_user_index = 1
  end
  return user
end

local function select_user()
  if has_invalid_users() and math.random() > valid_ratio then
    return next_invalid_user()
  end
  return next_user()
end

function init(args)
  configure_from_args(args)
  math.randomseed(os.time())

  wrk.method = "POST"
  wrk.headers["Content-Type"] = "application/json"
  wrk.headers["Accept"] = "application/json"
  wrk.headers["User-Agent"] = "wrk-login-bench/1.0"
end

function request()
  local user = select_user()
  wrk.path = path
  wrk.body = build_body(user.email, user.password)
  return wrk.format(nil, nil, nil, wrk.body)
end

function response(status, headers, body)
  if status ~= 200 and status ~= 401 and status ~= 429 then
    io.stderr:write(string.format("unexpected status=%d body=%s\n", status, body))
  end
end
