with recursive select_posts 
    AS (SELECT post_post.id, title, url, votes, created_at, username, (votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by rank desc limit 10),
    comments 
        AS ( 
            select post_post.id, select_posts.id as ancestor from post_post join select_posts on post_post.parent_id = select_posts.id where post_post.parent_id is not null 
            union all
            select post_post.id, ancestor from post_post join comments on post_post.parent_id = comments.id
        ),
    comments_count
        AS (
            select ancestor, count(*) as comment_count from comments group by ancestor
        ),
    user_votes 
        AS (
            select * from post_post_voters where post_post_voters.user_id=3
    ),
    user_flags
        AS (
            select * from post_post_flaggers where post_post_flaggers.user_id=3
        )
    select select_posts.*, user_votes.user_id as user_voted, comment_count, user_flags.user_id as user_flagged from select_posts left join user_votes on select_posts.id = user_votes.post_id left join comments_count on select_posts.id = comments_count.ancestor left join user_flags on select_posts.id = user_flags.post_id order by rank desc, created_at desc;

with recursive select_posts 
    AS (SELECT post_post.id, title, url, votes, created_at, username, (votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by rank desc limit 10)
    select post_post.id, select_posts.id as ancestor from post_post join select_posts on post_post.parent_id = select_posts.id where post_post.parent_id is not null;

select post_post.id, select_posts.id as ancestor from post_post join select_posts on post_post.parent_id = select_posts.id where post_post.parent_id is not null;

with recursive 
    comments 
        AS ( 
            select 0 as lev, *, SUBSTRING('0000' || id::VARCHAR, -4) || ' ' as skey from post_post where id=2 
            union all
            select lev + 1, post_post.*, skey || SUBSTRING('0000' || post_post.id::VARCHAR, -4) || ' ' from post_post join comments on post_post.parent_id = comments.id
        )
    select * from comments order by skey;