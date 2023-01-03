package field_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/pkg/v2/header"
	"github.com/zostay/go-email/pkg/v2/message"
)

func TestFoldValue(t *testing.T) {
	const emailMsg = `Delivered-To: sterling@example.com
Received: by 1.6.2.1 with SMTP id asdfasdfasdfasd;
        Fri, 30 Jan 2015 19:23:13 -0800 (PST)
X-Received: by 1.2.8.2 with SMTP id asdfasdfasdfasd.5.3;
        Fri, 30 Jan 2015 19:23:12 -0800 (PST)
Return-Path: <bounce-mc.us2_6.1-sterling=example.com@mail7.example.com>
Received: from mail7.example.com (mail7.example.com. [1.2.1.7])
        by mx.example.com with ESMTP id asdfasdfasdfasdf.1.2.0.3.1.2.1
        for <sterling@example.com>;
        Fri, 30 Jan 2015 19:23:12 -0800 (PST)
Received-SPF: pass (example.com: domain of bounce-mc.us2_6.1-sterling=example.com@mail7.example.com designates 1.2.1.7 as permitted sender) client-ip=1.2.1.7;
Authentication-Results: mx.example.com;
       spf=pass (example.com: domain of bounce-mc.us2_6.1-sterling=example.com@mail7.example.com designates 1.2.1.7 as permitted sender) smtp.mail=bounce-mc.us2_6.1-sterling=example.com@mail7.example.com;
       dkim=pass header.i=@mail7.example.com
DKIM-Signature: v=1; a=rsa-sha1; c=relaxed/relaxed; s=k1; d=mail7.example.com;
 h=Subject:From:Reply-To:To:Date:Message-ID:List-ID:List-Unsubscribe:Sender:Content-Type:MIME-Version; i=devsupport=3Dexample.com@mail7.example.com;
 bh=asdfasdfasdfasdfasdfasdfasdf;
 b=asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf
   asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf
   asdfasdfasdfasdfasdf
DomainKey-Signature: a=rsa-sha1; c=nofws; q=dns; s=k1; d=mail7.example.com;
 b=asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf
   asdfasfdasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf
   asdfasdfasdfasdfasdf;
Received: from (127.0.0.1) by mail7.example.com id asdfasdfasdf for <sterling@example.com>; Sat, 31 Jan 2015 03:23:09 +0000 (envelope-from <bounce-mc.us2_6.1-sterling=example.com@mail7.example.com>)
Subject: =?utf-8?Q?Emulator=20Behind=20The=20Scenes=2C=20Debugging=20Guides=2C=20New=20Meetup=20Groups=20and=20more=21?=
From: =?utf-8?Q?Example?= <devsupport@example.com>
Reply-To: =?utf-8?Q?Example?= <devsupport@gexample.com>
To: <sterling@example.com>
Date: Sat, 31 Jan 2015 03:23:09 +0000
Message-ID: <asdfasdfasdfasdfasdfasdfasdfasdfasd.2@mail7.example.com>
X-Mailer: MailChimp Mailer - **asdfasdfasdfasdfasdfasd**
X-Campaign: mailchimpasdfasdfasdfasdfasdfasdfa.asdfasdfas
X-campaignid: mailchimpasfdasdfasdfasdfasdfasdfa.asdfasdfas
X-Report-Abuse: Please report abuse for this campaign here: http://www.example.com/abuse/abuse.phtml?u=asdfasdfasdfasdfasdfasdfa&id=asdfasdfas&e=asdfasdfas
X-MC-User: asdfasdfasdfasdfasdfasdfa
X-Feedback-ID: 6:6.1:us2:mc
List-ID: asdfasdfasdfasdfasdfasdfasd list <asdfasdfasdfasdfasdfasdfa.7.list-id.example.com>
X-Accounttype: pd
List-Unsubscribe: <mailto:unsubscribe-asdfasdfasdfasdfasdfasdfa-asdfasdfas-asdfasdfas@mailin1.example.com?subject=unsubscribe>, <http://example.us2.example.com/unsubscribe?u=asdfasdfasdfasdfasdfasdfa&id=asdfasdfas&e=asdfasdfas&c=asdfasdfas>
Sender: "Example" <devsupport=example.com@mail7.example.com>
x-mcda: FALSE
Content-Type: multipart/alternative; boundary="_----------=_MCPart_433295335"
MIME-Version: 1.0
Keywords:

This is a multi-part message in MIME format

--_----------=_MCPart_433295335
Content-Type: text/plain; charset="utf-8"; format="fixed"
Content-Transfer-Encoding: quoted-printable

Hello.
--_----------=_MCPart_433295335--
`

	m, err := message.Parse(strings.NewReader(emailMsg), message.WithoutMultipart())
	assert.NoError(t, err)
	require.NotNil(t, m)

	assert.Equal(t, len(m.GetHeader().ListFields()), 30)

	from, err := m.GetHeader().Get(header.From)
	assert.NoError(t, err)
	assert.Equal(t, "Example <devsupport@example.com>", from)
}
