import { test, expect } from "../../src/fixtures/test";
import { requireData } from "../../src/api/devflow-api";
import { AppShell } from "../../src/pages/app-shell";
import { FeedPage } from "../../src/pages/feed.page";
import { NotificationsPage } from "../../src/pages/notifications.page";
import { PostDetailPage } from "../../src/pages/post-detail.page";
import { ProfilePage } from "../../src/pages/profile.page";
import { uniquePostTitle } from "../../src/support/data";

test.describe("UI｜跨用户链路", () => {
  test("关注开发者后，可以在关注流看到其动态", async ({ page, actor, data }) => {
    const author = await data.createActor();
    const ownedPost = await data.createPost(author, {
      title: uniquePostTitle("following")
    });
    const profile = new ProfilePage(page);
    await profile.goto(author.user.id);

    await profile.follow(author.user.id);
    data.deferApiCleanup("取消页面产生的关注关系", () => actor.api.unfollow(author.user.id));

    const state = requireData(await actor.api.followState(author.user.id), "查询关注状态");
    expect(state.followed).toBe(true);
    await new AppShell(page).openFollowing();
    await new FeedPage(page).expectPost(ownedPost.post.id, ownedPost.post);
  });

  test("作者可以查看点赞通知、标记已读并打开关联动态", async ({
    page,
    actor,
    data
  }) => {
    const ownedPost = await data.createPost(actor, {
      title: uniquePostTitle("notification")
    });
    const visitor = await data.createActor();
    await data.like(visitor, ownedPost);
    const notification = await data.waitForNotification(
      actor,
      (item) =>
        item.type === "like" &&
        item.actor_id === visitor.user.id &&
        item.post_id === ownedPost.post.id
    );
    const notifications = new NotificationsPage(page);
    const appShell = new AppShell(page);

    await notifications.goto();
    await appShell.expectUnreadCount(1);
    await notifications.expectText(
      notification.id,
      `${visitor.nickname} 点赞了 《${ownedPost.post.title}》`
    );
    await notifications.markRead(notification.id);
    await appShell.expectUnreadCount(0);
    await notifications.open(notification.id, ownedPost.post.id);
    await new PostDetailPage(page).expectPost(ownedPost.post.title);
  });
});
