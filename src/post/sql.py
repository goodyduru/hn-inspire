from collections import namedtuple
from django.db import connection

POST_LIMITS = 30
COMMENT_LIMITS = 200

authenticated_home_page_query = """with recursive posts 
    AS (SELECT post_post.id, title, url, votes, created_at, username, 
    (votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank 
    from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by rank desc limit %s),
    comments 
        AS ( 
            select post_post.id, posts.id as ancestor from post_post join posts on post_post.parent_id = posts.id where post_post.parent_id is not null 
            union all
            select post_post.id, ancestor from post_post join comments on post_post.parent_id = comments.id
        ),
    comments_count
        AS (
            select ancestor, count(*) as comment_count from comments group by ancestor
        ),
    user_votes 
        AS (
            select * from post_post_voters where post_post_voters.user_id=%s
    ),
    user_flags
        AS (
            select * from post_post_flaggers where post_post_flaggers.user_id=%s
        )
    select posts.id, posts.title, posts.url,
            posts.votes, posts.created_at, posts.username,
            user_votes.user_id as user_voted, comment_count, user_flags.user_id as user_flagged 
    from posts 
    left join user_votes on posts.id = user_votes.post_id 
    left join comments_count on posts.id = comments_count.ancestor 
    left join user_flags on posts.id = user_flags.post_id order by rank desc, created_at desc"""

authenticated_new_posts = """with recursive new_posts 
    AS (SELECT post_post.id, title, url, votes, created_at, username from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by created_at limit %s),
    comments 
        AS ( 
            select post_post.id, new_posts.id as ancestor from post_post join new_posts on post_post.parent_id = new_posts.id where post_post.parent_id is not null 
            union all
            select post_post.id, ancestor from post_post join comments on post_post.parent_id = comments.id
        ),
    comments_count
        AS (
            select ancestor, count(*) as comment_count from comments group by ancestor
        ),
    user_votes 
        AS (
            select * from post_post_voters where post_post_voters.user_id=%s
    ),
    user_flags
        AS (
            select * from post_post_flaggers where post_post_flaggers.user_id=%s
        )
    select new_posts.*, user_votes.user_id as user_voted, comment_count, user_flags.user_id as user_flagged 
    from new_posts 
    left join user_votes on new_posts.id = user_votes.post_id 
    left join comments_count on new_posts.id = comments_count.ancestor 
    left join user_flags on new_posts.id = user_flags.post_id order by created_at desc"""

unauthenticated_home_page_query = """with recursive no_user_home 
    AS (SELECT post_post.id, title, url, votes, created_at, username, (votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by rank desc limit %s),
    comments 
        AS ( 
            select post_post.id, no_user_home.id as ancestor from post_post join no_user_home on post_post.parent_id = no_user_home.id where post_post.parent_id is not null 
            union all
            select post_post.id, ancestor from post_post join comments on post_post.parent_id = comments.id
        ),
    comments_count
        AS (
            select ancestor, count(*) as comment_count from comments group by ancestor
        )
    select no_user_home.id, no_user_home.title, no_user_home.url,
            no_user_home.votes, no_user_home.created_at, no_user_home.username, comment_count 
    from no_user_home 
    left join comments_count on no_user_home.id = comments_count.ancestor order by rank desc, created_at desc"""

unauthenticated_new_posts = """with recursive no_user_new
    AS (SELECT post_post.id, title, url, votes, created_at, username from post_post join auth_user on post_post.author_id = auth_user.id where parent_id is null order by created_at limit %s),
    comments 
        AS ( 
            select post_post.id, no_user_new.id as ancestor from post_post join no_user_new on post_post.parent_id = no_user_new.id where post_post.parent_id is not null 
            union all
            select post_post.id, ancestor from post_post join comments on post_post.parent_id = comments.id
        ),
    comments_count
        AS (
            select ancestor, count(*) as comment_count from comments group by ancestor
        )
    select no_user_new.*, comment_count
    from no_user_new 
    left join comments_count on no_user_new.id = comments_count.ancestor order by created_at desc"""

authenticated_comments = """with recursive 
    comments_raw 
        AS ( 
            select 0 as lev, *, RIGHT('0000' || id::VARCHAR, 4) || ' ' as skey from post_post where id=%s
            union all
            select 
                lev + 1, 
                post_post.*, 
                skey || 
                RIGHT('0000' || (9999-post_post.votes)::VARCHAR, 4) || 
                RIGHT('0000000000' || (extract(epoch from (now() - post_post.created_at))::INTEGER)::VARCHAR, 10) || ' ' from post_post join comments_raw on post_post.parent_id = comments_raw.id
        ),
    comments
        AS (
            select * from comments_raw limit %s
        ),
    user_votes 
        AS (
            select * from post_post_voters where post_post_voters.user_id=%s
    ),
    user_flags
        AS (
            select * from post_post_flaggers where post_post_flaggers.user_id=%s
        )
    select 
        comments.id, comments.lev, 
        comments.title, comments.url, 
        comments.text, comments.votes,
        comments.created_at, comments.parent_id,
        username,
        user_votes.user_id as user_voted, user_flags.user_id as user_flagged
        from comments 
        join auth_user on comments.author_id = auth_user.id
        left join user_votes on comments.id = user_votes.post_id
        left join user_flags on comments.id = user_flags.post_id 
        order by skey"""

unauthenticated_comments = """with recursive 
    comments_raw 
        AS ( 
            select 0 as lev, *, RIGHT('0000' || id::VARCHAR, 4) || ' ' as skey from post_post where id=%s 
            union all
            select 
                lev + 1, 
                post_post.*, 
                skey || 
                RIGHT('0000' || (9999-post_post.votes)::VARCHAR, 4) || 
                RIGHT('0000000000' || (extract(epoch from (now() - post_post.created_at))::INTEGER)::VARCHAR, 10) || ' ' from post_post join comments_raw on post_post.parent_id = comments_raw.id
        ),
    comments
        AS (
            select * from comments_raw limit %s
        )
    select 
        comments.id, comments.lev, 
        comments.title, comments.url, 
        comments.text, comments.votes,
        comments.created_at, comments.parent_id,
        username
        from comments 
        join auth_user on comments.author_id = auth_user.id
        order by skey"""


def namedtuplefetchall(cursor, name):
    desc = cursor.description
    nt_result = namedtuple(name, [col[0] for col in desc])
    return [nt_result(*row) for row in cursor.fetchall()]


def get_home_page(user):
    with connection.cursor() as cursor:
        if user.is_authenticated:
            cursor.execute(
                authenticated_home_page_query, [POST_LIMITS, user.id, user.id]
            )
        else:
            cursor.execute(unauthenticated_home_page_query, [POST_LIMITS])
        results = namedtuplefetchall(cursor, "Post")
    return results


def get_comments(user, post_id):
    with connection.cursor() as cursor:
        if user.is_authenticated:
            cursor.execute(
                authenticated_comments, [post_id, COMMENT_LIMITS, user.id, user.id]
            )
        else:
            cursor.execute(unauthenticated_comments, [post_id, COMMENT_LIMITS])
        results = namedtuplefetchall(cursor, "Post")
    return results


def get_new_posts(user):
    with connection.cursor() as cursor:
        if user.is_authenticated:
            cursor.execute(authenticated_new_posts, [POST_LIMITS, user.id, user.id])
        else:
            cursor.execute(unauthenticated_new_posts, [POST_LIMITS])
        results = namedtuplefetchall(cursor, "Post")
    return results
