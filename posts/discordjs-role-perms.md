---
title: Make your Discord.js bot's global slash commands accessible only to certain roles
summary: Secrets that the documentation won't tell you (not sure why though).
time: 1631387096
---

So I was recently making a very simple [Discord bot](https://github.com/namansood/counter-bot) as a joke. Since I like JavaScript, I found a JavaScript library to help me out with this, [discord.js](https://discord.js.org), and got started.

As part of this, the bot had a slash command that was supposed to be accessible only to people with a certain role in the server. Now I imagine this situation comes up often (eg. only allow moderators to ban pepole, etc), so I was surprised that the documentation on this was a bit lacking. The main website is basically a reference page, and there's a [guide website](https://discordjs.guide) linked in the header. Using the guide, I was able to create the command that I wanted, minus the role thing.

There *is* a page on [slash command permissions](https://discordjs.guide/interactions/slash-command-permissions.html), which looked like exactly what I wanted, and which told me to effectively do this[^1]:

```js
// `client` is a Discord.js Client object
// name == `reset-counter` is the name of the slash command
await client.application.fetch();
const appCmd = await client.application.commands.fetch('<command id that I wanted to edit permissions for>');
const permissions = [
    {
        id: '<role id that I wanted to allow>',
        type: 'ROLE',
        permission: true
    }
];
appCmd.permissions.add(permissions);
```

You fetch the application details, you find the command you want, and you add the permissions you want to it. Simple enough, right? Except this doesn't work. Firstly, there's no clear way to get the ID of the slash command you want to edit permissions for. For that, it turns out you can just call `.commands.fetch()` without any arguments and it'll return a JavaScript collection of all the commands, indexed by ID. Then you can just loop through all the commands and look for the command with the right command name (in my case, `reset-counter`). Secondly, even after fetching the right command and adding permissions to it, it would fail with the following error:

```
Permissions for global commands may only be fetched or modified by providing a GuildResolvable or from a guild's application command manager.
```

Notice that I'm getting my command from `client.application`, not `client.guilds.fetch('<discord server id>')`, because it is a global command, not a guild command. This is because I don't want my bot to be restricted to a single Discord server/guild, I want it to run in multiple servers. The guide doesn't mention a way of providing a GuildResolvable, but I can see from Discord.js's lovely type annotations that `appCmd.permissions` is an `ApplicationCommandPermissionsManager`. Now that I have a class name, I can just look it up in the reference!

[Here](https://discord.js.org/#/docs/discord.js/stable/class/ApplicationCommandPermissionsManager) is the reference. The `add()` method, according to this, takes an `AddApplicationCommandPermissionsOptions`, which is an array of `ApplicationCommandPermissionData` -- the exact same thing that `permissions` in the code above is. Oh well.

When documentation fails me, I like to look at the source code to see if I can get any hints there. [Here](https://github.com/discordjs/discord.js/blob/988a51b7641f8b33cc9387664605ddc02134859d/src/managers/ApplicationCommandPermissionsManager.js#L222) is the beginning of the `add()` method:

```js
async add({ guild, command, permissions }) {
    const { guildId, commandId } = this._validateOptions(guild, command);
    ...
}
```

Wait a second. In addition to the permissions array, `add()` also takes the command ID to edit permissions for, and the guild ID of the relevant guild! This is undocumented, for some reason. The comment over the add method mentions the command ID, but not the guild ID. There is no mention that I saw anywhere on the internet that suggested that `add()` would take the guild ID as a parameter. But now that I know it does, editing the last line to:

```js
appCmd.permissions.add({
    permissions: permissions,
    guildId: '<id of the server I want the permissions to apply in>'
});
```

works fine!

Sadly, the guild ID is defined _outside_ the permissions array, so if I want to add permissions for different sets of servers, I need to make multiple calls to `add()`. The final code ended up looking like this:

```js
// find commands by name instead of ID
appCommands.filter(appCmd => appCmd.name === name).forEach(async appCmd => {
    for(perm of cmd.permissions) {
        await appCmd.permissions.add(perm);
    }
});
```

Where each `perm` is an object that looks like this:

```js
{ 
    guild: '<id of the server>',
    permissions: [
        {
            id: '<id of the role or user to allow>',
            type: '"ROLE" or "USER"',
            permission: true,
        }
    ],
}
```

And that works! The bot is not allowed to be used by anyone except those people defined in `cmd.permissions`, across multiple servers. The one downside is that it doesn't work in DMs, but I *think* it might work there on a per-user basis if I make a `perm` object for the user and don't mention a guild ID. I haven't tried it, but feel free to do so.

[^1]: Note that while creating your slash command, you also need to call `.setDefaultPermission(false);` on your `SlashCommandBuilder`. This effectively makes your bot rejects commands from everyone, and what you pass to `command.permissions.add()` becomes your allowlist.
