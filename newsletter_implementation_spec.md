# 📰 Especificação Técnica - Implementação de Newsletters (Fase 1)

## 🎯 **Objetivo**
Implementar os 6 métodos básicos de newsletters do WhatsApp baseado na biblioteca whatsmeow, seguindo o padrão arquitetural do zpwoot.

## 📋 **Métodos a Implementar**

### **🔥 Alta Prioridade (6 métodos)**
1. `CreateNewsletter` - Criar canal/newsletter
2. `GetNewsletterInfo` - Informações do canal
3. `GetNewsletterInfoWithInvite` - Info do canal via convite
4. `FollowNewsletter` - Seguir canal
5. `UnfollowNewsletter` - Deixar de seguir canal
6. `GetSubscribedNewsletters` - Listar canais seguidos

## 🏗️ **Estrutura de Implementação**

### **1. Entidades e DTOs**

#### **1.1 Newsletter Entity (`internal/domain/newsletter/entity.go`)**
```go
package newsletter

import (
    "errors"
    "time"
)

// Domain errors
var (
    ErrInvalidNewsletterJID = errors.New("invalid newsletter JID")
    ErrInvalidNewsletterName = errors.New("invalid newsletter name")
    ErrNewsletterNameTooLong = errors.New("newsletter name too long")
    ErrDescriptionTooLong = errors.New("description too long")
    ErrInvalidInviteKey = errors.New("invalid invite key")
    ErrNewsletterNotFound = errors.New("newsletter not found")
    ErrNotNewsletterAdmin = errors.New("user is not a newsletter admin")
)

// NewsletterInfo representa um canal do WhatsApp
type NewsletterInfo struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    InviteCode      string                 `json:"inviteCode"`
    SubscriberCount int                    `json:"subscriberCount"`
    State           string                 `json:"state"`
    Role            string                 `json:"role"`
    Muted           bool                   `json:"muted"`
    Verified        bool                   `json:"verified"`
    CreationTime    time.Time              `json:"creationTime"`
    Picture         *ProfilePictureInfo    `json:"picture,omitempty"`
}

// CreateNewsletterRequest representa os dados para criar um canal
type CreateNewsletterRequest struct {
    Name        string `json:"name" validate:"required,max=25"`
    Description string `json:"description,omitempty" validate:"max=512"`
    Picture     []byte `json:"picture,omitempty"`
}

// ProfilePictureInfo representa informações da foto do canal
type ProfilePictureInfo struct {
    URL    string `json:"url"`
    ID     string `json:"id"`
    Type   string `json:"type"`
    Direct string `json:"direct"`
}
```

#### **1.2 Newsletter DTOs (`internal/app/newsletter/dto.go`)**
```go
package newsletter

import "time"

// CreateNewsletterRequest - Request para criar newsletter
type CreateNewsletterRequest struct {
    Name        string `json:"name" validate:"required,max=25"`
    Description string `json:"description,omitempty" validate:"max=512"`
}

// CreateNewsletterResponse - Response da criação de newsletter
type CreateNewsletterResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    InviteCode  string    `json:"inviteCode"`
    CreatedAt   time.Time `json:"createdAt"`
}

// GetNewsletterInfoRequest - Request para obter info de newsletter
type GetNewsletterInfoRequest struct {
    JID string `json:"jid" validate:"required"`
}

// GetNewsletterInfoWithInviteRequest - Request para obter info via convite
type GetNewsletterInfoWithInviteRequest struct {
    InviteKey string `json:"inviteKey" validate:"required"`
}

// NewsletterInfoResponse - Response com informações do newsletter
type NewsletterInfoResponse struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    InviteCode      string                 `json:"inviteCode"`
    SubscriberCount int                    `json:"subscriberCount"`
    State           string                 `json:"state"`
    Role            string                 `json:"role"`
    Muted           bool                   `json:"muted"`
    Verified        bool                   `json:"verified"`
    CreationTime    time.Time              `json:"creationTime"`
    Picture         *ProfilePictureInfo    `json:"picture,omitempty"`
}

// FollowNewsletterRequest - Request para seguir newsletter
type FollowNewsletterRequest struct {
    JID string `json:"jid" validate:"required"`
}

// UnfollowNewsletterRequest - Request para deixar de seguir newsletter
type UnfollowNewsletterRequest struct {
    JID string `json:"jid" validate:"required"`
}

// SubscribedNewslettersResponse - Response com newsletters seguidos
type SubscribedNewslettersResponse struct {
    Newsletters []NewsletterInfoResponse `json:"newsletters"`
    Total       int                      `json:"total"`
}

// NewsletterActionResponse - Response genérica para ações
type NewsletterActionResponse struct {
    JID       string    `json:"jid"`
    Status    string    `json:"status"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
```

### **2. WameowClient Extensions**

#### **2.1 Métodos no WameowClient (`internal/infra/wameow/client.go`)**
```go
// CreateNewsletter creates a new WhatsApp channel
func (c *WameowClient) CreateNewsletter(ctx context.Context, name, description string) (*types.NewsletterMetadata, error) {
    if !c.client.IsLoggedIn() {
        return nil, fmt.Errorf("client is not logged in")
    }

    params := whatsmeow.CreateNewsletterParams{
        Name:        name,
        Description: description,
    }

    c.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
        "session_id":  c.sessionID,
        "name":        name,
        "description": description,
    })

    newsletter, err := c.client.CreateNewsletter(params)
    if err != nil {
        c.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
            "session_id": c.sessionID,
            "error":      err.Error(),
        })
        return nil, err
    }

    c.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
        "session_id":    c.sessionID,
        "newsletter_id": newsletter.ID.String(),
        "name":          name,
    })

    return newsletter, nil
}

// GetNewsletterInfo gets information about a newsletter
func (c *WameowClient) GetNewsletterInfo(ctx context.Context, jid string) (*types.NewsletterMetadata, error) {
    if !c.client.IsLoggedIn() {
        return nil, fmt.Errorf("client is not logged in")
    }

    parsedJID, err := c.parseJID(jid)
    if err != nil {
        return nil, fmt.Errorf("invalid JID: %w", err)
    }

    c.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
        "session_id": c.sessionID,
        "jid":        jid,
    })

    newsletter, err := c.client.GetNewsletterInfo(parsedJID)
    if err != nil {
        c.logger.ErrorWithFields("Failed to get newsletter info", map[string]interface{}{
            "session_id": c.sessionID,
            "jid":        jid,
            "error":      err.Error(),
        })
        return nil, err
    }

    return newsletter, nil
}

// GetNewsletterInfoWithInvite gets newsletter info using invite key
func (c *WameowClient) GetNewsletterInfoWithInvite(ctx context.Context, inviteKey string) (*types.NewsletterMetadata, error) {
    if !c.client.IsLoggedIn() {
        return nil, fmt.Errorf("client is not logged in")
    }

    c.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
        "session_id": c.sessionID,
        "invite_key": inviteKey,
    })

    newsletter, err := c.client.GetNewsletterInfoWithInvite(inviteKey)
    if err != nil {
        c.logger.ErrorWithFields("Failed to get newsletter info with invite", map[string]interface{}{
            "session_id": c.sessionID,
            "invite_key": inviteKey,
            "error":      err.Error(),
        })
        return nil, err
    }

    return newsletter, nil
}

// FollowNewsletter follows a newsletter
func (c *WameowClient) FollowNewsletter(ctx context.Context, jid string) error {
    if !c.client.IsLoggedIn() {
        return fmt.Errorf("client is not logged in")
    }

    parsedJID, err := c.parseJID(jid)
    if err != nil {
        return fmt.Errorf("invalid JID: %w", err)
    }

    c.logger.InfoWithFields("Following newsletter", map[string]interface{}{
        "session_id": c.sessionID,
        "jid":        jid,
    })

    err = c.client.FollowNewsletter(parsedJID)
    if err != nil {
        c.logger.ErrorWithFields("Failed to follow newsletter", map[string]interface{}{
            "session_id": c.sessionID,
            "jid":        jid,
            "error":      err.Error(),
        })
        return err
    }

    c.logger.InfoWithFields("Newsletter followed successfully", map[string]interface{}{
        "session_id": c.sessionID,
        "jid":        jid,
    })

    return nil
}

// UnfollowNewsletter unfollows a newsletter
func (c *WameowClient) UnfollowNewsletter(ctx context.Context, jid string) error {
    if !c.client.IsLoggedIn() {
        return fmt.Errorf("client is not logged in")
    }

    parsedJID, err := c.parseJID(jid)
    if err != nil {
        return fmt.Errorf("invalid JID: %w", err)
    }

    c.logger.InfoWithFields("Unfollowing newsletter", map[string]interface{}{
        "session_id": c.sessionID,
        "jid":        jid,
    })

    err = c.client.UnfollowNewsletter(parsedJID)
    if err != nil {
        c.logger.ErrorWithFields("Failed to unfollow newsletter", map[string]interface{}{
            "session_id": c.sessionID,
            "jid":        jid,
            "error":      err.Error(),
        })
        return err
    }

    c.logger.InfoWithFields("Newsletter unfollowed successfully", map[string]interface{}{
        "session_id": c.sessionID,
        "jid":        jid,
    })

    return nil
}

// GetSubscribedNewsletters gets all subscribed newsletters
func (c *WameowClient) GetSubscribedNewsletters(ctx context.Context) ([]*types.NewsletterMetadata, error) {
    if !c.client.IsLoggedIn() {
        return nil, fmt.Errorf("client is not logged in")
    }

    c.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
        "session_id": c.sessionID,
    })

    newsletters, err := c.client.GetSubscribedNewsletters()
    if err != nil {
        c.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
            "session_id": c.sessionID,
            "error":      err.Error(),
        })
        return nil, err
    }

    c.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
        "session_id": c.sessionID,
        "count":      len(newsletters),
    })

    return newsletters, nil
}
```

### **3. Endpoints da API**

#### **3.1 Rotas (`internal/infra/http/routers/routes.go`)**
```go
// Newsletter management routes
newsletterHandler := handlers.NewNewsletterHandler(appLogger, container.GetNewsletterUseCase(), container.GetSessionRepository())
sessions.Post("/:sessionId/newsletters/create", newsletterHandler.CreateNewsletter)                    // POST /sessions/:sessionId/newsletters/create
sessions.Get("/:sessionId/newsletters/info", newsletterHandler.GetNewsletterInfo)                      // GET /sessions/:sessionId/newsletters/info?jid=...
sessions.Post("/:sessionId/newsletters/info-from-invite", newsletterHandler.GetNewsletterInfoWithInvite) // POST /sessions/:sessionId/newsletters/info-from-invite
sessions.Post("/:sessionId/newsletters/follow", newsletterHandler.FollowNewsletter)                    // POST /sessions/:sessionId/newsletters/follow
sessions.Post("/:sessionId/newsletters/unfollow", newsletterHandler.UnfollowNewsletter)                // POST /sessions/:sessionId/newsletters/unfollow
sessions.Get("/:sessionId/newsletters", newsletterHandler.GetSubscribedNewsletters)                    // GET /sessions/:sessionId/newsletters
```

#### **3.2 Exemplos de Uso da API**

**Criar Newsletter:**
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/create" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Meu Canal",
    "description": "Descrição do meu canal"
  }'
```

**Obter Info de Newsletter:**
```bash
curl -X GET "http://localhost:8080/sessions/SESSION_ID/newsletters/info?jid=120363123456789012@newsletter" \
  -H "Authorization: ZP_API_KEY"
```

**Seguir Newsletter:**
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/follow" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "jid": "120363123456789012@newsletter"
  }'
```

**Listar Newsletters Seguidos:**
```bash
curl -X GET "http://localhost:8080/sessions/SESSION_ID/newsletters" \
  -H "Authorization: ZP_API_KEY"
```

## 🔄 **Fluxo de Implementação**

### **Ordem de Implementação:**
1. ✅ Criar estrutura base (entidades, DTOs, interfaces)
2. ✅ Implementar métodos no WameowClient
3. ✅ Criar Newsletter UseCase
4. ✅ Implementar Newsletter Handler
5. ✅ Configurar rotas
6. ✅ Testes e validação

### **Padrões a Seguir:**
- **Arquitetura**: Clean Architecture (Domain, Application, Infrastructure)
- **Logging**: Usar logger estruturado com campos contextuais
- **Validação**: Usar tags de validação nos DTOs
- **Errors**: Retornar erros específicos do domínio
- **Testes**: Cobertura mínima de 80%

## 📝 **Notas Técnicas**

### **Tipos WhatsApp Newsletter JID:**
- Formato: `120363123456789012@newsletter`
- Diferente de grupos: `@g.us`
- Diferente de usuários: `@s.whatsapp.net`

### **Validações Importantes:**
- Nome do newsletter: máximo 25 caracteres
- Descrição: máximo 512 caracteres
- JID deve ter formato válido de newsletter
- Invite key deve ser válida

### **Tratamento de Erros:**
- Cliente não logado
- JID inválido
- Newsletter não encontrado
- Permissões insuficientes
- Erros de rede/timeout

## 🎯 **Resultado Esperado**

Após a implementação, o zpwoot terá suporte completo aos newsletters básicos do WhatsApp, permitindo:
- ✅ Criar canais
- ✅ Obter informações de canais
- ✅ Seguir/deixar de seguir canais
- ✅ Listar canais seguidos
- ✅ Obter informações via convite

Isso representa **50% da funcionalidade de newsletters** (6 de 12 métodos), focando nos recursos mais importantes e utilizados.
