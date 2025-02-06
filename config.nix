{
  nixos,
  # list representing a nixos option path (e.g. ['console' 'enable']), or a
  # prefix of such a path (e.g. ['console']), or a string representing the same
  # (e.g. 'console.enable')
  path,
  # whether to recurse down the config attrset and show each set value instead
  recursive,
}: let
  configEntry = lib.attrByPath path' null nixos.config;
in
  if !lib.hasAttrByPath path' nixos.config
  then throw "Couldn't resolve config path '${lib.showOption path'}'"
  else let
    optionEntry = optionByPath path' nixos.options;
    configEntry = lib.attrByPath path' null nixos.config;
  in
    if recursive
    then renderRecursive configEntry
    else renderFull optionEntry configEntry
