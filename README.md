# 🥊 Torneira de Soco Inglês

**Offensive Security Toolkit** - Uma suíte modular de pentest desenvolvida em Go para testes de segurança.

```
╔═══════════════════════════════════════════════════════════════════╗
║   ████████╗ ██████╗ ██████╗ ███╗   ██╗███████╗██╗██████╗  █████╗  ║
║   ╚══██╔══╝██╔═══██╗██╔══██╗████╗  ██║██╔════╝██║██╔══██╗██╔══██╗ ║
║      ██║   ██║   ██║██████╔╝██╔██╗ ██║█████╗  ██║██████╔╝███████║ ║
║      ██║   ██║   ██║██╔══██╗██║╚██╗██║██╔══╝  ██║██╔══██╗██╔══██║ ║
║      ██║   ╚██████╔╝██║  ██║██║ ╚████║███████╗██║██║  ██║██║  ██║ ║
║      ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ║
║                   	 DE SOCO INGLÊS                          	║
║              [ Offensive Security Toolkit v1.0 ]              	║
╚═══════════════════════════════════════════════════════════════════╝
```
## Instalação

### Pré-requisitos
- Go 1.21+

### Compilar
```bash
git clone https://github.com/seu-usuario/torneira-de-soco-ingles.git
cd torneira-de-soco-ingles
go build -o skynet.exe .
```

```bash
# Scan nas portas mais comuns
skynet scan -t 192.168.1.1 --top

# Scan em todas as portas (1-65535)
skynet scan -t 192.168.1.1 --all

# Scan em portas específicas
skynet scan -t 192.168.1.1 -p 80,443,8080

# Scan em range de portas
skynet scan -t 192.168.1.1 -p 1-1000

# Salvar resultado em arquivo
skynet scan -t 192.168.1.1 --top -o resultado.txt
```

**Flags:**
| Flag | Descrição |
|------|-----------|
| `-t, --target` | Host alvo (IP ou domínio) |
| `-p, --ports` | Portas específicas ou range |
| `--top` | Portas mais comuns |
| `--all` | Todas as portas (1-65535) |
| `--threads` | Número de threads (default: 100) |
| `--timeout` | Timeout em segundos (default: 2) |
| `-o, --output` | Arquivo de saída |

---

### 2. Enumeração (`enum`)

Descobre diretórios, subdomínios e virtual hosts.

```bash
# Enumerar diretórios
skynet enum -t http://target.com -m dir -w wordlists/dirs.txt

# Enumerar com extensões específicas
skynet enum -t http://target.com -m dir -w wordlists/dirs.txt -x php,html,txt

# Enumerar subdomínios
skynet enum -t target.com -m sub -w wordlists/subdomains.txt

# Enumerar virtual hosts
skynet enum -t http://192.168.1.1 -m vhost -w wordlists/subdomains.txt

# Salvar resultados
skynet enum -t http://target.com -m dir -w wordlists/dirs.txt -o found.txt
```

**Flags:**
| Flag | Descrição |
|------|-----------|
| `-t, --target` | URL ou domínio alvo |
| `-m, --mode` | Modo: `dir`, `sub`, `vhost` |
| `-w, --wordlist` | Wordlist para enumeração |
| `-x, --extension` | Extensões (ex: php,html,txt) |
| `--threads` | Número de threads (default: 30) |
| `-o, --output` | Arquivo de saída |

---

### 3. Força Bruta (`brute`)

Ataque de força bruta em formulários de login.

```bash
# Brute force com usuário específico
skynet brute -t http://target.com/login -u admin -P wordlists/passwords.txt \
  --user-field username --pass-field password --fail "Invalid"

# Brute force com lista de usuários
skynet brute -t http://target.com/login -U wordlists/users.txt -P wordlists/passwords.txt \
  --user-field user --pass-field pass --success "Welcome"

# Com delay entre requisições (evitar bloqueio)
skynet brute -t http://target.com/login -u admin -P wordlists/passwords.txt \
  --user-field username --pass-field password --fail "Invalid" --delay 500
```

**Flags:**
| Flag | Descrição |
|------|-----------|
| `-t, --target` | URL do formulário de login |
| `-u, --username` | Usuário específico |
| `-U, --userlist` | Wordlist de usuários |
| `-P, --passlist` | Wordlist de senhas |
| `--user-field` | Nome do campo de usuário |
| `--pass-field` | Nome do campo de senha |
| `--success` | String que indica sucesso |
| `--fail` | String que indica falha |
| `--threads` | Número de threads (default: 10) |
| `--delay` | Delay entre requisições (ms) |

---

### 4. Fuzzing (`fuzz`)

Fuzzing de parâmetros para descoberta de vulnerabilidades.

```bash
# Fuzzing de diretórios
skynet fuzz -t http://target.com/FUZZ -w wordlists/dirs.txt

# Fuzzing com payloads de SQL Injection
skynet fuzz -t "http://target.com/page?id=FUZZ" --payload sqli

# Fuzzing com payloads de XSS
skynet fuzz -t "http://target.com/search?q=FUZZ" --payload xss

# Fuzzing POST
skynet fuzz -t http://target.com/login -m POST -d "user=admin&pass=FUZZ" -w wordlists/passwords.txt

# Filtrar códigos de status (ignorar 404 e 403)
skynet fuzz -t http://target.com/FUZZ -w wordlists/dirs.txt --fc 404,403

# Mostrar apenas códigos específicos
skynet fuzz -t http://target.com/FUZZ -w wordlists/dirs.txt --mc 200,301
```

**Tipos de Payload Disponíveis:**
| Tipo | Descrição |
|------|-----------|
| `sqli` | SQL Injection |
| `xss` | Cross-Site Scripting |
| `lfi` | Local File Inclusion |
| `rfi` | Remote File Inclusion |
| `cmdi` | Command Injection |
| `ssti` | Server-Side Template Injection |
| `ssrf` | Server-Side Request Forgery |
| `xxe` | XML External Entity |
| `nosqli` | NoSQL Injection |
| `all` | Todos os payloads |

**Flags:**
| Flag | Descrição |
|------|-----------|
| `-t, --target` | URL com marcador `FUZZ` |
| `-w, --wordlist` | Wordlist customizada |
| `--payload` | Tipo de payload integrado |
| `-m, --method` | Método HTTP (GET/POST) |
| `-d, --data` | Dados POST |
| `-H, --header` | Headers customizados |
| `--fc` | Filtrar códigos de status |
| `--mc` | Mostrar apenas códigos |
| `--fs` | Filtrar por tamanho |
| `--threads` | Threads (default: 20) |

---

### 5. Exploit (`exploit`)

Testa vulnerabilidades conhecidas automaticamente.

```bash
# Testar SQL Injection
skynet exploit -t "http://target.com/page?id=1" --type sqli

# Testar XSS
skynet exploit -t "http://target.com/search?q=test" --type xss

# Testar LFI
skynet exploit -t "http://target.com/page?file=index" --type lfi

# Testar todas as vulnerabilidades
skynet exploit -t "http://target.com/page?id=1" --all

# Testar em parâmetro específico
skynet exploit -t "http://target.com/page" --type sqli -p id

# Modo verboso
skynet exploit -t "http://target.com/page?id=1" --all -v
```

**Flags:**
| Flag | Descrição |
|------|-----------|
| `-t, --target` | URL alvo |
| `--type` | Tipo de vulnerabilidade |
| `--all` | Testar todas |
| `-p, --param` | Parâmetro específico |
| `-m, --method` | Método HTTP |
| `-v, --verbose` | Modo verboso |

---

## Exemplos Práticos

### Reconhecimento Completo
```bash
# 1. Scan de portas
skynet scan -t 192.168.1.100 --top -o ports.txt

# 2. Enumerar diretórios
skynet enum -t http://192.168.1.100 -m dir -w wordlists/dirs.txt -o dirs.txt

# 3. Testar vulnerabilidades encontradas
skynet exploit -t "http://192.168.1.100/admin?id=1" --all
```

### Ataque a Login
```bash
# 1. Identificar campos do formulário (inspecionar HTML)
# 2. Executar brute force
skynet brute -t http://target.com/login \
  -u admin \
  -P wordlists/passwords.txt \
  --user-field username \
  --pass-field password \
  --fail "Invalid credentials" \
  --threads 5 \
  --delay 100
```

Este projeto é apenas para fins educacionais.

