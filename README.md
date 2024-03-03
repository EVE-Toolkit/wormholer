# wormholer 🕳️
a Discord bot for EVE wormhole corps

- make buy & sell requests for trade hub runs
  - request fits & items
  - get notified when your request is fulfilled

<a href="https://ibb.co/LZ3KtJG"><img src="https://i.ibb.co/RDq80B5/Frame-1-1.png" alt="Frame-1-1" border="0"></a>

- make signature reports
  - sigs are automatically parsed & formatted
  - used so your corp knows what's been scanned & unscanned
 
<a href="https://ibb.co/jWZV2Vt"><img src="https://i.ibb.co/WPpDwDC/Screenshot-2024-03-02-210608.png" alt="Screenshot-2024-03-02-210608" border="0"></a>

## configuration

1. configure your .env to look like the following:
```
TOKEN=discord bot token
```

## scan reporting
```
$system=system name
copy paste sigs from probe scanner
```

## buy request
```
$buy item
```
OR
```
$buy
item(s) newlines are created with shift + enter
```

## sell request
```
$sell hangar=corp hangar name items=items to sell
```
