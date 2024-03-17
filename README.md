# distributed-caching-and-loadbalancing-system
A basic implementation and visualization of caching and load balancing system for distributed platform. It's a college project that implements basic implementation of data structures like LinkedLists, Hash Maps and Pointers.

```image
      _,---.      _,.---._                   .--, .-.--,                  .=-.-. ,--.--------.                                              
  _.='.'-,  \   ,-.' , -  `.                 |  |=| -\==\                /==/_ //==/,  -   , -\          _.-.      .-.,.---.   .--.-. .-.-. 
 /==.'-     /  /==/_,  ,  - \  ,--.--------. |  `-' _|==| ,--.--------. |==|, | \==\.-.  - ,-./        .-,.'|     /==/  `   \ /==/ -|/=/  | 
/==/ -   .-'  |==|   .=.     |/==/,  -   , -\\     , |==|/==/,  -   , -\|==|  |  `--`\==\- \          |==|, |    |==|-, .=., ||==| ,||=| -| 
|==|_   /_,-. |==|_ : ;=:  - |\==\.-.  - ,-./ `--.  -|==|\==\.-.  - ,-./|==|- |       \==\_ \         |==|- |    |==|   '='  /|==|- | =/  | 
|==|  , \_.' )|==| , '='     | `--`--------`      \_ |==| `--`--------` |==| ,|       |==|- |         |==|, |    |==|- ,   .' |==|,  \/ - | 
\==\-  ,    (  \==\ -    ,_ /                     |  \==\               |==|- |       |==|, |         |==|- `-._ |==|_  . ,'. |==|-   ,   / 
 /==/ _  ,  /   '.='. -   .'                       \ /==/               /==/. /       /==/ -/         /==/ - , ,//==/  /\ ,  )/==/ , _  .'  
 `--`------'      `--`--''                          `--`                `--`-`        `--`--`         `--`-----' `--`-`--`--' `--`..---'    

```

## Project Structure

```
distributed-caching-and-loadbalancing-system/
├── caching/
│   ├── cache/
│   │   ├── cache.go          // Cache implementation
│   │   ├── cacher.go         // Cache interface
│   │   ├── command.go        // Command processing logic
│   │   └── persist.go        // AOF persistence logic
│   └── replication.go        // Replication logic
│
├── server/
│   ├── config.go 
│   ├── master.go             // Master server implementation
│   ├── slave.go              // Slave server implementation
│   └── server-node.go 
│         
├── loadbalancer/
│   └── loadbalancer.go  
│
├── client/
│   └── client.go 
│
├── tmp/                       // Temporary files
│   └── aof.log
│
├── config.yml                 // Configuration file
└── main.go                    // Main application logic

```